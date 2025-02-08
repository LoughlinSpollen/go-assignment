package network

import (
	"assignment/lib/shared_lib/domain"
	"assignment/lib/shared_lib/dto"
	"assignment/lib/shared_lib/transport"
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type Publisher interface {
	Connect(ctx context.Context, queueName, queueConnStr string) error
	Publish(ctx context.Context, cmd *domain.Command) ([]domain.DataEntry, error)
	Close() error
}

type resp struct {
	dataEntries []domain.DataEntry
	err         error
	doneC       chan struct{}
}

type publisher struct {
	dtoAdapter        dto.DTOAdapter
	connector         transport.MQConnector
	conn              transport.MQConnection
	channel           transport.MQChannel
	serviceQueueName  string
	replyQueueName    string
	respMap           sync.Map // map[correlationID]*resp
	closeOnce         sync.Once
	stopListeningChan chan struct{}
}

func NewPublisher(dtoAdapter dto.DTOAdapter, connector transport.MQConnector) Publisher {
	return &publisher{
		dtoAdapter: dtoAdapter,
		connector:  connector,
		respMap:    sync.Map{},
	}
}

func (pub *publisher) Connect(ctx context.Context, serviceQueueName, queueConnStr string) error {
	log.Trace("publisher Connect")

	conn, err := pub.connector.Dial(queueConnStr)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("rabbitMQ connect error")
		wrappedErr := fmt.Errorf("rabbitMQ connect error: %w", err)
		return wrappedErr
	}
	pub.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		log.WithContext(ctx).WithError(err).Error("rabbitMQ channel open error")
		wrappedErr := fmt.Errorf("rabbitMQ channel open error: %w", err)
		return wrappedErr

	}
	pub.channel = ch

	replyQueueName := fmt.Sprintf("%s_reply_%s", serviceQueueName, uuid.NewString())
	q, err := pub.channel.QueueDeclare(
		replyQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		log.WithContext(ctx).WithError(err).Error("rabbitMQ declare queue failed")
		wrappedErr := fmt.Errorf("rabbitMQ declare queue failed, queue name is %s : %w", q.Name, err)
		return wrappedErr
	}

	respCh, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		log.WithContext(ctx).WithError(err).Error("rabbitMQ consumer channel failed")
		wrappedErr := fmt.Errorf("rabbitMQ consumer channel failed for queue %s : %w", q.Name, err)
		return wrappedErr
	}
	log.Info("opened queue: ", q.Name)

	pub.conn = conn
	pub.channel = ch
	pub.serviceQueueName = serviceQueueName
	pub.replyQueueName = q.Name
	pub.stopListeningChan = make(chan struct{})
	go pub.waitForResponse(ctx, respCh)

	return nil
}

func (pub *publisher) Publish(ctx context.Context, cmd *domain.Command) ([]domain.DataEntry, error) {
	log.Trace("publisher Publish")

	correlationID := uuid.NewString()
	respMsg := &resp{
		doneC: make(chan struct{}),
	}
	pub.respMap.Store(correlationID, respMsg)

	body, err := pub.dtoAdapter.ToDTO(ctx, cmd)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to convert command to DTO")
		pub.respMap.Delete(correlationID)
		wrappedErr := fmt.Errorf("failed to convert command to DTO: %w", err)
		return nil, wrappedErr
	}

	reqMsg := amqp091.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationID,
		ReplyTo:       pub.replyQueueName,
		Body:          body,
	}

	err = pub.channel.PublishWithContext(
		ctx,
		"",
		pub.serviceQueueName,
		false,
		false,
		reqMsg,
	)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to publish message")
		pub.respMap.Delete(correlationID)
		wrappedErr := fmt.Errorf("failed to publish message: %w", err)
		return nil, wrappedErr
	}

	select {
	case <-respMsg.doneC:
		return respMsg.dataEntries, respMsg.err
	case <-ctx.Done():
		// Context cancelled / time-out
		pub.respMap.Delete(correlationID)
		return nil, ctx.Err()
	}
}

func (pub *publisher) waitForResponse(ctx context.Context, delivery <-chan amqp091.Delivery) {
	log.Trace("publisher waitForResponse")

	for {
		select {
		case <-pub.stopListeningChan:
			log.Info("publisher listener stopped")
			return
		case d, ok := <-delivery:
			if !ok {
				log.Info("publisher listener channel closed")
				return
			}
			correlationID := d.CorrelationId
			val, ok := pub.respMap.Load(correlationID)
			if !ok {
				// dangling response
				continue
			}
			respMsg := val.(*resp)

			statusVal, _ := d.Headers["x-status"]
			switch statusVal {
			case "error":
				var err error
				respMsg.err, err = pub.dtoAdapter.FromErrorDTO(ctx, d.Body)
				if err != nil {
					log.WithContext(ctx).WithError(err).Error("failed to convert error DTO to error")
					respMsg.err = fmt.Errorf("failed to read server error: %w", err)
				}
			default:
				cmd, err := pub.dtoAdapter.FromDTO(ctx, d.Body)
				if err != nil {
					log.WithContext(ctx).WithError(err).Error("failed to convert DTO to command")
					respMsg.err = fmt.Errorf("failed to read server response: %w", err)
					continue
				}
				respMsg.dataEntries = cmd.Data
			}

			close(respMsg.doneC)
			pub.respMap.Delete(correlationID)
		}
	}
}

func (pub *publisher) Close() error {
	var err error
	pub.closeOnce.Do(func() {
		close(pub.stopListeningChan)
		err = pub.channel.Close()
		if err != nil {
			log.WithError(err).Error("failed to close channel")
			_ = pub.conn.Close()
		}
		err = pub.conn.Close()
		if err != nil {
			log.WithError(err).Error("failed to close connection")
		}
	})
	return err
}
