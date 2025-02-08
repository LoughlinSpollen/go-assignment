package network_test

import (
	"assignment/lib/shared_lib/domain"
	"assignment/lib/shared_lib/transport"
	"assignment_service/pkg/infra/network"

	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
)

const (
	testConnStr      = "test-conn"
	testQueueNameStr = "test-queue"
)

var (
	cmd = domain.Command{
		Name: domain.AddItem,
		Data: []domain.DataEntry{{Key: "testKey", Value: "testValue"}},
	}
	cmdAsBytes, _ = json.Marshal(cmd)
	queueMsg      = amqp091.Delivery{
		Body:          cmdAsBytes,
		ReplyTo:       "test-reply",
		ContentType:   "application/json",
		CorrelationId: "test-correlation",
	}
)

func TestSubscriber(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Subscriber test suite")
}

type mockMQConnector struct {
	ReturnedMockConnection transport.MQConnection
	ReturnedDialErr        error
	DialCalled             bool
	ExpectedConnStr        string
}

func (m *mockMQConnector) Dial(connectionString string) (transport.MQConnection, error) {
	Expect(connectionString).To(Equal(m.ExpectedConnStr))
	m.DialCalled = true
	return m.ReturnedMockConnection, m.ReturnedDialErr
}

type mockMQConnection struct {
	ReturnedMockChannel transport.MQChannel
	ReturnedChannelErr  error
	ReturnedCloseErr    error
	ChannelCalled       bool
	CloseCalled         bool
}

func (m *mockMQConnection) Channel() (transport.MQChannel, error) {
	m.ChannelCalled = true
	return m.ReturnedMockChannel, m.ReturnedChannelErr
}

func (m *mockMQConnection) Close() error {
	m.CloseCalled = true
	return m.ReturnedCloseErr
}

type mockMQChannel struct {
	ReturnedQueueMsg         amqp091.Delivery
	ReturnedQueueDeclareErr  error
	ReturnedQueueCloseErr    error
	ReturnedQueuePublishErr  error
	ReturnedQueueConsumeErr  error
	ReturnedDelcareQueueName string
	QueueDeclareCalled       bool
	ChannelCloseCalled       bool
	QueuePublishCalled       bool
	QueueConsumeCalled       bool
	ExpectedQueueName        string
	ExpectedPublshedMsg      amqp091.Publishing
	ConsumeWg                *sync.WaitGroup
	PublishWg                *sync.WaitGroup
	MsgConsumed              bool
}

func (m *mockMQChannel) QueueDeclare(name string, _, _, _, _ bool, _ amqp091.Table) (amqp091.Queue, error) {
	Expect(name).To(Equal(m.ExpectedQueueName))
	m.QueueDeclareCalled = true
	return amqp091.Queue{Name: m.ReturnedDelcareQueueName}, m.ReturnedQueueDeclareErr
}

func (m *mockMQChannel) Close() error {
	m.ChannelCloseCalled = true
	return nil
}

func (m *mockMQChannel) PublishWithContext(_ context.Context, _, _ string, _, _ bool, msg amqp091.Publishing) error {
	Expect(msg.ContentType).To(Equal(m.ExpectedPublshedMsg.ContentType))
	Expect(msg.Headers).To(Equal(m.ExpectedPublshedMsg.Headers))
	Expect(msg.Body).To(Equal(m.ExpectedPublshedMsg.Body))

	m.QueuePublishCalled = true
	if m.PublishWg != nil {
		m.PublishWg.Done()
		m.PublishWg = nil
	}
	return m.ReturnedQueuePublishErr
}

func (m *mockMQChannel) Consume(queue, _ string, _, _, _, _ bool, _ amqp091.Table) (<-chan amqp091.Delivery, error) {
	Expect(queue).To(Equal(m.ExpectedQueueName))
	m.QueueConsumeCalled = true
	var msgs chan amqp091.Delivery
	if m.ReturnedQueueMsg.Body != nil && !m.MsgConsumed {
		msgs = make(chan amqp091.Delivery, 1)
		go func() {
			msgs <- m.ReturnedQueueMsg
			close(msgs)
		}()
		m.MsgConsumed = true
		return msgs, m.ReturnedQueueConsumeErr
	} else {
		// returning nil will deadlock, can't signal close
		msgs := make(chan amqp091.Delivery)
		close(msgs)
	}

	if m.ConsumeWg != nil {
		m.ConsumeWg.Done()
		m.ConsumeWg = nil
	}
	return msgs, m.ReturnedQueueConsumeErr
}

type mockDTOAdapter struct {
	ReturnedToDTOError       error
	ReturnedFromDTOError     error
	ReturnedDTOBytes         []byte
	ReturnedDTOCommand       *domain.Command
	ReturnedToErrDTOBytes    []byte
	ReturnedToErrDTOError    error
	ReturnedFromErrDTOResult error
	ToDTOCalled              bool
	FromDTOCalled            bool
	ToErrDTOCalled           bool
	FromErrDTOCalled         bool
	ExpectedCmd              *domain.Command
	ExpectedDTOBytes         []byte
	ExpectedDTOError         error
	ExpectErrDTOBytes        []byte
	FromDTOWg                *sync.WaitGroup
}

func (m *mockDTOAdapter) ToDTO(_ context.Context, cmd *domain.Command) ([]byte, error) {
	for k, v := range m.ExpectedCmd.Data {
		Expect(cmd.Data[k]).To(Equal(v))
	}
	Expect(cmd.Name).To(Equal(m.ExpectedCmd.Name))
	m.ToDTOCalled = true
	return m.ReturnedDTOBytes, m.ReturnedToDTOError
}

func (m *mockDTOAdapter) FromDTO(_ context.Context, cmdBytes []byte) (*domain.Command, error) {
	Expect(cmdBytes).To(Equal(m.ExpectedDTOBytes))
	m.FromDTOCalled = true
	if m.FromDTOWg != nil {
		m.FromDTOWg.Done()
		m.FromDTOWg = nil
	}
	return m.ReturnedDTOCommand, m.ReturnedFromDTOError
}

func (m *mockDTOAdapter) RegisterCommandTypeValidation() error {
	return nil
}

func (m *mockDTOAdapter) ToErrorDTO(_ context.Context, err error) ([]byte, error) {
	Expect(err.Error()).To(ContainSubstring(m.ExpectedDTOError.Error()))
	m.ToErrDTOCalled = true
	return m.ReturnedToErrDTOBytes, m.ReturnedToErrDTOError
}

func (m *mockDTOAdapter) FromErrorDTO(_ context.Context, errBytes []byte) (error, error) {
	Expect(errBytes).To(Equal(m.ExpectErrDTOBytes))
	m.FromErrDTOCalled = true
	return m.ReturnedFromErrDTOResult, m.ReturnedFromErrDTOResult
}

type mockUsecase struct {
	ReturnedHandleError error
	HandleCalled        bool
	ExpectedCmd         *domain.Command
	UsecaseWg           *sync.WaitGroup
}

func (m *mockUsecase) HandleCommand(ctx context.Context, cmd *domain.Command) error {
	for k, v := range m.ExpectedCmd.Data {
		Expect(cmd.Data[k]).To(Equal(v))
	}
	Expect(cmd.Name).To(Equal(m.ExpectedCmd.Name))
	m.HandleCalled = true
	if m.UsecaseWg != nil {
		m.UsecaseWg.Done()
		m.UsecaseWg = nil
	}
	return m.ReturnedHandleError
}

var _ = Describe("Subscriber", func() {
	var (
		mockDTO        *mockDTOAdapter
		mockCmdUsecase *mockUsecase
		mockConn       *mockMQConnection
		mockCh         *mockMQChannel
		mockConnector  *mockMQConnector
		subscriber     network.Subscriber
		cancelCtx      context.Context
		cancelFtn      context.CancelFunc
	)

	BeforeEach(func() {
		mockDTO = &mockDTOAdapter{}
		mockCmdUsecase = &mockUsecase{}
		mockCh = &mockMQChannel{}
		mockConn = &mockMQConnection{ReturnedMockChannel: mockCh}
		mockConnector = &mockMQConnector{ReturnedMockConnection: mockConn}
		cancelCtx, cancelFtn = context.WithCancel(context.Background())
		mockConnector.ExpectedConnStr = testConnStr
		mockConnector.ReturnedMockConnection = mockConn
		mockConn.ReturnedMockChannel = mockCh
		mockCh.ReturnedDelcareQueueName = testQueueNameStr
		mockCh.ExpectedQueueName = testQueueNameStr
		subscriber = network.NewSubscriber(testConnStr, testQueueNameStr, mockDTO, mockCmdUsecase, mockConnector)
	})

	Context("Positive test cases", func() {
		It("should Connect successfully with a valid connection", func() {
			var wg sync.WaitGroup
			wg.Add(1)
			mockCh.ConsumeWg = &wg

			go func() {
				err := subscriber.Connect(cancelCtx)
				Expect(err).NotTo(HaveOccurred())
			}()

			wg.Wait()
			cancelFtn()

			Expect(mockConnector.DialCalled).To(BeTrue())
			Expect(mockConn.ChannelCalled).To(BeTrue())
			Expect(mockCh.QueueDeclareCalled).To(BeTrue())
			Expect(mockCh.QueueConsumeCalled).To(BeTrue())

			subscriber.Close()
		})

		It("should process a message successfully and publish the result", func() {
			mockDTO.ReturnedDTOCommand = &cmd
			mockDTO.ExpectedCmd = &cmd
			mockCmdUsecase.ExpectedCmd = &cmd

			mockCh.ReturnedQueueMsg = queueMsg
			mockDTO.ReturnedDTOBytes = cmdAsBytes
			mockDTO.ExpectedDTOBytes = cmdAsBytes
			mockCh.ExpectedPublshedMsg = amqp091.Publishing{
				ContentType: "application/json",
				Headers:     amqp091.Table{"x-status": "ok"},
				Body:        cmdAsBytes,
			}

			var wg sync.WaitGroup
			wg.Add(1)
			mockCh.PublishWg = &wg

			go func() {
				err := subscriber.Connect(cancelCtx)
				Expect(err).NotTo(HaveOccurred())
			}()

			wg.Wait()
			cancelFtn()

			Expect(mockDTO.FromDTOCalled).To(BeTrue())
			Expect(mockCmdUsecase.HandleCalled).To(BeTrue())
			Expect(mockDTO.ToDTOCalled).To(BeTrue())
			Expect(mockCh.QueuePublishCalled).To(BeTrue())

			subscriber.Close()
		})
	})

	Context("Negative test cases", func() {
		It("should fail if Dial fails", func() {
			mockConnector.ReturnedDialErr = errors.New("dial failure")
			err := subscriber.Connect(cancelCtx)
			Expect(err).To(HaveOccurred())
			Expect(mockConnector.DialCalled).To(BeTrue())

			subscriber.Close()
		})

		It("should fail if Channel creation fails", func() {
			mockConn.ReturnedChannelErr = errors.New("channel failure")
			err := subscriber.Connect(cancelCtx)
			Expect(err).To(HaveOccurred())
			Expect(mockConn.ChannelCalled).To(BeTrue())
			Expect(mockCmdUsecase.HandleCalled).To(BeFalse())
			subscriber.Close()
		})

		It("should fail if Queue declaration fails", func() {
			mockCh.ReturnedQueueDeclareErr = errors.New("queue declaration failure")
			err := subscriber.Connect(cancelCtx)
			Expect(err).To(HaveOccurred())
			Expect(mockCh.QueueDeclareCalled).To(BeTrue())
			Expect(mockDTO.FromDTOCalled).To(BeFalse())
			Expect(mockCmdUsecase.HandleCalled).To(BeFalse())

			subscriber.Close()
		})

		It("should fail if DTO unmarshaling fails and return a bad request error", func() {
			mockCh.ReturnedQueueMsg = queueMsg
			mockDTO.ExpectedDTOBytes = cmdAsBytes
			mockDTO.ReturnedFromDTOError = errors.New("unmarshal failure")
			mockDTO.ExpectedDTOError = errors.New("Bad request error")
			mockDTO.ReturnedToErrDTOBytes, _ = json.Marshal(map[string]string{"Message": mockDTO.ExpectedDTOError.Error()})
			mockCh.ExpectedPublshedMsg = amqp091.Publishing{
				ContentType: "application/json",
				Headers:     amqp091.Table{"x-status": "error"},
				Body:        mockDTO.ReturnedToErrDTOBytes,
			}

			var wg sync.WaitGroup
			wg.Add(1)
			mockDTO.FromDTOWg = &wg

			go func() {
				err := subscriber.Connect(cancelCtx)
				Expect(err).NotTo(HaveOccurred())
			}()

			wg.Wait()
			cancelFtn()

			Expect(mockCh.QueueDeclareCalled).To(BeTrue())
			Expect(mockDTO.FromDTOCalled).To(BeTrue())
			Expect(mockDTO.ToErrDTOCalled).To(BeTrue())
			Expect(mockCmdUsecase.HandleCalled).To(BeFalse())
			subscriber.Close()
		})

		It("should fail if Usecase fails to handle command and return an error", func() {
			mockCh.ReturnedQueueMsg = queueMsg
			mockDTO.ExpectedDTOBytes = cmdAsBytes
			mockDTO.ReturnedDTOCommand = &cmd
			mockCmdUsecase.ExpectedCmd = &cmd
			mockCmdUsecase.ReturnedHandleError = errors.New("usecase failure")
			mockDTO.ExpectedDTOError = errors.New("Internal server error: failed to handle command")
			mockDTO.ReturnedToErrDTOBytes, _ = json.Marshal(map[string]string{"Message": mockDTO.ExpectedDTOError.Error()})
			mockCh.ExpectedPublshedMsg = amqp091.Publishing{
				ContentType: "application/json",
				Headers:     amqp091.Table{"x-status": "error"},
				Body:        mockDTO.ReturnedToErrDTOBytes,
			}

			var wg sync.WaitGroup
			wg.Add(1)
			mockCmdUsecase.UsecaseWg = &wg

			go func() {
				err := subscriber.Connect(cancelCtx)
				Expect(err).NotTo(HaveOccurred())
			}()

			wg.Wait()
			cancelFtn()

			Expect(mockCh.QueueDeclareCalled).To(BeTrue())
			Expect(mockDTO.FromDTOCalled).To(BeTrue())
			Expect(mockCmdUsecase.HandleCalled).To(BeTrue())
			subscriber.Close()
		})
	})
})
