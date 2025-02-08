package network_test

import (
	"assignment/lib/shared_lib/domain"
	"assignment/lib/shared_lib/transport"
	"assignment_client/pkg/infra/network"
	"context"
	"encoding/json"
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

	queueMsg = amqp091.Delivery{
		Body:        cmdAsBytes,
		ReplyTo:     "test-reply",
		ContentType: "application/json",
	}
)

func TestPublisher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Publisher test suite")
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

	QueueDeclareCalled bool
	ChannelCloseCalled bool
	QueuePublishCalled bool
	QueueConsumeCalled bool
	MsgConsumed        bool

	ExpectedServerQueueName string
	ExpectedPublshedMsg     amqp091.Publishing

	msgChan       chan amqp091.Delivery
	publishCalled bool
}

func (m *mockMQChannel) QueueDeclare(
	name string, _, _, exclusive, _ bool, _ amqp091.Table,
) (amqp091.Queue, error) {
	Expect(name).To(ContainSubstring(m.ExpectedServerQueueName))
	Expect(name).To(ContainSubstring("_reply_"))
	m.QueueDeclareCalled = true
	return amqp091.Queue{Name: m.ReturnedDelcareQueueName}, m.ReturnedQueueDeclareErr
}

func (m *mockMQChannel) Close() error {
	m.ChannelCloseCalled = true
	return m.ReturnedQueueCloseErr
}

func (m *mockMQChannel) PublishWithContext(
	_ context.Context, _, _ string,
	_, _ bool,
	msg amqp091.Publishing,
) error {
	Expect(msg.ContentType).To(Equal(m.ExpectedPublshedMsg.ContentType))
	Expect(msg.ReplyTo).To(Equal(m.ExpectedPublshedMsg.ReplyTo))
	Expect(msg.Body).To(Equal(m.ExpectedPublshedMsg.Body))

	m.QueuePublishCalled = true
	m.publishCalled = true

	if m.msgChan != nil && m.ReturnedQueueMsg.Body != nil && !m.MsgConsumed {
		go func() {
			m.ReturnedQueueMsg.CorrelationId = msg.CorrelationId
			m.msgChan <- m.ReturnedQueueMsg
		}()
		m.MsgConsumed = true
	}

	return m.ReturnedQueuePublishErr
}

func (m *mockMQChannel) Consume(
	queue, _ string, _, _, _, _ bool,
	_ amqp091.Table,
) (<-chan amqp091.Delivery, error) {
	Expect(queue).To(Equal(m.ExpectedServerQueueName))
	m.QueueConsumeCalled = true

	msgs := make(chan amqp091.Delivery, 1)
	m.msgChan = msgs

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
	return m.ReturnedDTOCommand, m.ReturnedFromDTOError
}

func (m *mockDTOAdapter) RegisterCommandTypeValidation() error {
	return nil
}

func (m *mockDTOAdapter) ToErrorDTO(_ context.Context, err error) ([]byte, error) {
	Expect(err).To(Equal(m.ExpectedDTOError))
	m.ToErrDTOCalled = true
	return m.ReturnedToErrDTOBytes, m.ReturnedToErrDTOError
}

func (m *mockDTOAdapter) FromErrorDTO(_ context.Context, errBytes []byte) (error, error) {
	Expect(errBytes).To(Equal(m.ExpectErrDTOBytes))
	m.FromErrDTOCalled = true
	return m.ReturnedFromErrDTOResult, m.ReturnedFromErrDTOResult
}

var _ = Describe("Publisher", func() {
	var (
		mockDTO       *mockDTOAdapter
		mockConn      *mockMQConnection
		mockCh        *mockMQChannel
		mockConnector *mockMQConnector
		publisher     network.Publisher
		ctx           context.Context
	)

	BeforeEach(func() {
		mockDTO = &mockDTOAdapter{}
		mockCh = &mockMQChannel{}
		mockConn = &mockMQConnection{ReturnedMockChannel: mockCh}
		mockConnector = &mockMQConnector{ReturnedMockConnection: mockConn}

		ctx = context.Background()
		mockConnector.ExpectedConnStr = testConnStr
		mockConnector.ReturnedMockConnection = mockConn
		mockConn.ReturnedMockChannel = mockCh

		mockCh.ReturnedDelcareQueueName = testQueueNameStr
		mockCh.ExpectedServerQueueName = testQueueNameStr
		publisher = network.NewPublisher(mockDTO, mockConnector)
	})

	Context("Positive test cases", func() {

		It("should Connect successfully with a valid connection", func() {
			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockConnector.DialCalled).To(BeTrue())
			Expect(mockConn.ChannelCalled).To(BeTrue())
			Expect(mockCh.QueueDeclareCalled).To(BeTrue())
			Expect(mockCh.QueueConsumeCalled).To(BeTrue())

			err = publisher.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should publish successfully with a valid connection", func() {
			mockDTO.ExpectedCmd = &cmd
			mockDTO.ReturnedDTOBytes = cmdAsBytes
			mockCh.ExpectedPublshedMsg = amqp091.Publishing{
				ContentType: "application/json",
				ReplyTo:     testQueueNameStr,
				Body:        cmdAsBytes,
			}
			mockCh.ReturnedQueueMsg = queueMsg
			mockDTO.ReturnedDTOCommand = &cmd
			mockDTO.ExpectedDTOBytes = cmdAsBytes

			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)
			Expect(err).NotTo(HaveOccurred())
			dataEntries, err := publisher.Publish(ctx, &cmd)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockDTO.ToDTOCalled).To(BeTrue())
			Expect(mockCh.QueuePublishCalled).To(BeTrue())
			Expect(len(dataEntries)).To(Equal(len(mockDTO.ReturnedDTOCommand.Data)))
			for k, v := range dataEntries {
				Expect(v).To(Equal(mockDTO.ReturnedDTOCommand.Data[k]))
			}

			err = publisher.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should Close successfully with a valid connection", func() {
			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)
			Expect(err).NotTo(HaveOccurred())

			err = publisher.Close()
			Expect(err).NotTo(HaveOccurred())
			Expect(mockCh.ChannelCloseCalled).To(BeTrue())
			Expect(mockConn.CloseCalled).To(BeTrue())

			err = publisher.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Negative test cases", func() {
		It("should return an error when Connect fails to dial", func() {
			mockConnector.ReturnedDialErr = amqp091.ErrClosed
			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)

			Expect(err).To(HaveOccurred())
			Expect(mockConnector.DialCalled).To(BeTrue())
			Expect(mockConn.ChannelCalled).To(BeFalse())
			Expect(mockCh.QueueDeclareCalled).To(BeFalse())
			Expect(mockCh.QueueConsumeCalled).To(BeFalse())
		})

		It("should return an error when Connect fails to create a channel", func() {
			mockConn.ReturnedChannelErr = amqp091.ErrClosed
			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)

			Expect(err).To(HaveOccurred())
			Expect(mockConnector.DialCalled).To(BeTrue())
			Expect(mockConn.ChannelCalled).To(BeTrue())
			Expect(mockCh.QueueDeclareCalled).To(BeFalse())
			Expect(mockCh.QueueConsumeCalled).To(BeFalse())
		})

		It("should return an error when Connect fails to declare a queue", func() {
			mockCh.ReturnedQueueDeclareErr = amqp091.ErrClosed
			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)

			Expect(err).To(HaveOccurred())
			Expect(mockConnector.DialCalled).To(BeTrue())
			Expect(mockConn.ChannelCalled).To(BeTrue())
			Expect(mockCh.QueueDeclareCalled).To(BeTrue())
			Expect(mockCh.QueueConsumeCalled).To(BeFalse())
		})

		It("should return an error when Connect fails to consume a queue", func() {
			mockCh.ReturnedQueueConsumeErr = amqp091.ErrClosed
			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)

			Expect(err).To(HaveOccurred())
			Expect(mockConnector.DialCalled).To(BeTrue())
			Expect(mockConn.ChannelCalled).To(BeTrue())
			Expect(mockCh.QueueDeclareCalled).To(BeTrue())
			Expect(mockCh.QueueConsumeCalled).To(BeTrue())
		})

		It("should return an error when Publish fails to publish a message", func() {
			err := publisher.Connect(ctx, testQueueNameStr, testConnStr)
			Expect(err).NotTo(HaveOccurred())

			mockCh.ReturnedQueuePublishErr = amqp091.ErrClosed
			mockDTO.ExpectedCmd = &cmd
			mockDTO.ReturnedDTOBytes = cmdAsBytes
			mockCh.ExpectedPublshedMsg = amqp091.Publishing{
				ContentType: "application/json",
				ReplyTo:     testQueueNameStr,
				Body:        cmdAsBytes,
			}

			dataEntries, err := publisher.Publish(ctx, &cmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to publish message"))
			Expect(mockCh.QueuePublishCalled).To(BeTrue())
			Expect(len(dataEntries)).To(Equal(0))

			err = publisher.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
