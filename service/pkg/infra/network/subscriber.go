package network

import (
	"assignment/lib/shared_lib/dto"
	"assignment_service/pkg/app"
	"context"
	"fmt"
	"time"

	"assignment/lib/shared_lib/transport"

	"github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	exclusiveQueue = false
	autoDelete     = false
	durable        = false
	noWait         = false
)

type contextKey string

type Subscriber interface {
	Connect(ctx context.Context) error
	Close()
}

type subscriber struct {
	queueConnStr string
	queueName    string
	cmdUsecase   app.Usecase
	dtoAdapter   dto.DTOAdapter

	conntr  transport.MQConnector
	conn    transport.MQConnection
	channel transport.MQChannel
}

func NewSubscriber(queueConnStr, queueName string, dtoAdapter dto.DTOAdapter, cmdUsecase app.Usecase, conntr transport.MQConnector) *subscriber {
	log.Trace("new subscriber")

	return &subscriber{
		queueConnStr: queueConnStr,
		queueName:    queueName,
		cmdUsecase:   cmdUsecase,
		dtoAdapter:   dtoAdapter,
		conntr:       conntr,
	}
}

func (sub *subscriber) Connect(ctx context.Context) error {
	log.Trace("subscriber Connect")

	conn, err := sub.conntr.Dial(sub.queueConnStr)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("rabbitMQ connect error")
		wrappedErr := fmt.Errorf("rabbitMQ connect error: %w", err)
		return wrappedErr
	}
	sub.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		log.WithContext(ctx).WithError(err).Error("rabbitMQ channel open error")
		wrappedErr := fmt.Errorf("rabbitMQ channel open error: %w", err)
		return wrappedErr
	}
	sub.channel = channel

	_, err = channel.QueueDeclare(
		sub.queueName,
		durable,
		autoDelete,
		exclusiveQueue,
		noWait,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		_ = conn.Close()
		log.WithContext(ctx).WithError(err).Error("rabbitMQ declare queue failed")
		wrappedErr := fmt.Errorf("rabbitMQ declare queue failed, queue name is %s : %w", sub.queueName, err)
		return wrappedErr
	}

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		return sub.consumeLoop(ctx)
	})
	return errs.Wait()
}

func (sub *subscriber) Close() {
	log.Trace("subscriber Close")

	if sub.channel != nil {
		err := sub.channel.Close()
		if err != nil {
			log.WithError(err).Error("rabbitMQ channel close error")
		}
	}
	if sub.conn != nil {
		err := sub.conn.Close()
		if err != nil {
			log.WithError(err).Error("rabbitMQ connection close error")
		}
	}
}

func (sub *subscriber) consumeLoop(ctx context.Context) error {
	log.Trace("subscriber consumeLoop")

	log.Info("subscriber polling loop Connected")
	defer func() {
		log.Info("subscriber polling loop stopped")
	}()

	var errResult error

	run := true
	for run {
		select {
		case <-ctx.Done():
			run = false
		default:
			msgs, err := sub.channel.Consume(
				sub.queueName,
				"", // Consumer name auto-generated
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				// TODO: exponential backoff reconnect
				errResult := fmt.Errorf("rabbitMQ consume error: %w", err)
				log.WithContext(ctx).WithError(errResult).Error("rabbitMQ consume error")
				ctx.Done()
			}

			for m := range msgs {
				go sub.processMessage(m)
			}
		}
	}

	return errResult
}

func (sub *subscriber) processMessage(m amqp091.Delivery) {
	log.Trace("subscriber processMessage")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if m.CorrelationId != "" {
		ctx = context.WithValue(ctx, contextKey("correlation_id"), m.CorrelationId)
	}

	if m.ReplyTo == "" {
		log.Warn("replyTo field is empty. Cannot reply to client")
		return
	}

	if m.ContentType != "application/json" {
		replyErr := fmt.Errorf("Bad request error: invalid content type")
		log.Warn(replyErr)
		sub.replyWithError(ctx, m, replyErr)
		return
	}

	if m.Body == nil {
		replyErr := fmt.Errorf("Bad request error: empty message body")
		log.Warn(replyErr)
		sub.replyWithError(ctx, m, replyErr)
		return
	}

	cmd, err := sub.dtoAdapter.FromDTO(ctx, m.Body)
	if err != nil {
		// TODO: add to dead letter queue - for now return error
		replyErr := fmt.Errorf("Bad request error: failed to unmarshal message. %s", err.Error())
		log.Warn(err)
		sub.replyWithError(ctx, m, replyErr)
		return
	}

	err = sub.cmdUsecase.HandleCommand(ctx, cmd)
	if err != nil {
		log.Warn(err)
		sub.replyWithError(ctx, m, err)
		return
	}

	dtoBytes, err := sub.dtoAdapter.ToDTO(ctx, cmd)
	if err != nil {
		replyErr := fmt.Errorf("Internal server error: failed to marshal response message")
		log.WithError(err).Error(replyErr.Error())
		sub.replyWithError(ctx, m, replyErr)
		return
	}

	log.Info("replying to client: ", m.ReplyTo)
	err = sub.channel.PublishWithContext(
		ctx,
		"",
		m.ReplyTo,
		false,
		false,
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: m.CorrelationId,
			Headers: amqp091.Table{
				"x-status": "ok",
			},
			Body: dtoBytes,
		},
	)
	if err != nil {
		// TODO: Expodential backoff retry
		replyErr := fmt.Errorf("Internal server error: failed to publish response message")
		log.WithError(err).Error(replyErr.Error())
		sub.replyWithError(ctx, m, replyErr)
		return
	}
}

func (sub *subscriber) replyWithError(ctx context.Context, m amqp091.Delivery, err error) {
	log.Trace("subscriber replyWithError")

	errJson, err := sub.dtoAdapter.ToErrorDTO(ctx, err)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal error message. Cannot reply to client")
		return
	}

	err = sub.channel.PublishWithContext(
		ctx,
		"",
		m.ReplyTo,
		false,
		false,
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: m.CorrelationId,
			Headers: amqp091.Table{
				"x-status": "error",
			},
			Body: errJson,
		},
	)
	if err != nil {
		log.WithError(err).Fatal("failed to publish error message. Cannot reply to client")
	}
}
