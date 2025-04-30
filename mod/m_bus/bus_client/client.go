// file:buzz/mod/m_bus/bus_client/client.go

package bus_client

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/rskv-p/buzz/mod/m_bus/bus_req"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

var _ typ.IBusClient = (*BusClient)(nil)

//-----------------------------------------
//  BusClient Struct
//-----------------------------------------

// BusClient handles client-side message bus operations.
type BusClient struct {
	ID                uint64
	Bus               typ.IBus
	Subscriptions     *sync.Map
	Client            typ.IBusClient
	messageChannel    chan typ.IRequest
	secretKey         string
	batchBuffer       []typ.IRequest
	batchTicker       *time.Ticker
	mu                sync.Mutex
	cond              *sync.Cond
	msgHandlers       map[string]func(subject string, msg []byte)
	maxConcurrentMsgs int
	Connection        net.Conn
}

//-----------------------------------------
//  Initialization
//-----------------------------------------

// NewBusClient initializes a new BusClient instance.
func NewBusClient(id uint64, bus typ.IBus, secretKey string, maxConcurrentMsgs int, host string, port int) typ.IBusClient {
	address := formatAddress(host, port)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		x_log.Error("Error connecting to bus:", err)
		return nil
	}

	client := &BusClient{
		ID:                id,
		Bus:               bus,
		Subscriptions:     &sync.Map{},
		messageChannel:    make(chan typ.IRequest, 100),
		secretKey:         secretKey,
		batchBuffer:       make([]typ.IRequest, 0),
		batchTicker:       time.NewTicker(5 * time.Second),
		msgHandlers:       make(map[string]func(subject string, msg []byte)),
		maxConcurrentMsgs: maxConcurrentMsgs,
		Connection:        conn,
	}
	client.cond = sync.NewCond(&client.mu)

	go client.processMessages()
	x_log.Info("BusClient", id, "created and ready to process messages")
	return client
}

// formatAddress prepares TCP address based on IP version.
func formatAddress(host string, port int) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}

//-----------------------------------------
//  Message Processing
//-----------------------------------------

// RegisterHandler sets a handler function for a specific subject.
func (bc *BusClient) RegisterHandler(subject string, handler func(subject string, msg []byte)) error {
	if bc.msgHandlers == nil {
		bc.msgHandlers = make(map[string]func(subject string, msg []byte))
	}
	bc.msgHandlers[subject] = handler
	x_log.Info("Handler registered for subject", subject, "on client", bc.ID)
	return nil
}

// processMessages listens on messageChannel and processes messages asynchronously.
func (bc *BusClient) processMessages() {
	x_log.Info("Client", bc.ID, "started processing messages")

	sem := make(chan struct{}, bc.maxConcurrentMsgs)

	for {
		select {
		case req := <-bc.messageChannel:
			sem <- struct{}{}
			go func(req typ.IRequest) {
				defer func() { <-sem }()
				if request, ok := req.(*bus_req.Request); ok {
					x_log.Info("Client", bc.ID, "received message for subject:", request.GetSubject())
					bc.batchBuffer = append(bc.batchBuffer, request)
					if len(bc.batchBuffer) >= 100 {
						bc.handleBatch(bc.batchBuffer)
						bc.batchBuffer = nil
					}
				} else {
					x_log.Error("Client", bc.ID, "received invalid message type")
				}
			}(req)
		case <-bc.batchTicker.C:
			if len(bc.batchBuffer) > 0 {
				bc.handleBatch(bc.batchBuffer)
				bc.batchBuffer = nil
			}
		}
	}
}

// handleBatch processes a slice of incoming requests.
func (bc *BusClient) handleBatch(batch []typ.IRequest) {
	for _, req := range batch {
		subject := req.GetSubject()
		data := req.GetData()
		handler, exists := bc.GetHandler(subject)
		if !exists {
			x_log.Error("Client", bc.ID, "No handler found for subject", subject)
			continue
		}
		handler(subject, data)
	}
}

//-----------------------------------------
//  Subscriptions
//-----------------------------------------

// Subscribe subscribes to a subject/queue with a message handler.
func (bc *BusClient) Subscribe(subject, queue string, handler func(subject string, msg []byte)) error {
	return bc.subscribeTopic(subject, queue, handler)
}

// RetrySubscribe attempts to subscribe with retries and delays.
func (bc *BusClient) RetrySubscribe(subject, queue string, retries int, delay time.Duration, handler func(subject string, msg []byte)) error {
	for i := 0; i < retries; i++ {
		err := bc.Subscribe(subject, queue, handler)
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("client %d: failed to subscribe after %d attempts", bc.ID, retries)
}

// subscribeTopic is an internal method that manages topic/queue subscriptions.
func (bc *BusClient) subscribeTopic(subject, queue string, handler func(subject string, msg []byte)) error {
	subscriptions, _ := bc.Subscriptions.LoadOrStore(subject, []string{})
	existingQueues := subscriptions.([]string)

	for _, existingQueue := range existingQueues {
		if existingQueue == queue {
			return nil
		}
	}

	bc.Subscriptions.Store(subject, append(existingQueues, queue))

	if err := bc.RegisterHandler(subject, handler); err != nil {
		return fmt.Errorf("failed to register handler: %w", err)
	}

	if err := bc.Bus.Subscribe([]byte(subject), []byte(queue), bc); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	return nil
}

//-----------------------------------------
//  Publishing
//-----------------------------------------

// PublishMessage sends a message to the bus.
func (bc *BusClient) PublishMessage(subject string, data []byte, queue string) error {
	if err := bc.Bus.Publish([]byte(subject), []byte(queue), data); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// RetryPublish attempts to send a message with retries and backoff.
func (bc *BusClient) RetryPublish(subject string, data []byte, retries int, delay time.Duration, queue string) error {
	currentDelay := delay
	for i := 0; i < retries; i++ {
		if err := bc.PublishMessage(subject, data, queue); err == nil {
			return nil
		}
		time.Sleep(currentDelay)
		currentDelay *= 2
	}
	return fmt.Errorf("client %d: failed to publish message after %d attempts", bc.ID, retries)
}

//-----------------------------------------
//  Request Handling
//-----------------------------------------

// SendRequest processes a request through middleware and sends it.
func (bc *BusClient) SendRequest(ctx context.Context, req typ.IRequest) error {
	for _, middleware := range bc.Bus.GetMiddleware() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := middleware.Process(req); err != nil {
				return err
			}
		}
	}
	return nil
}

//-----------------------------------------
//  Internal Message Routing
//-----------------------------------------

// SendToMessageChannel enqueues a message to the processing channel.
func (bc *BusClient) SendToMessageChannel(subject string, data []byte) error {
	select {
	case bc.messageChannel <- &bus_req.Request{Subject: subject, Data: data}:
		return nil
	default:
		return fmt.Errorf("message channel is full")
	}
}

// HandleIncomingMessage receives and logs an incoming message.
func (bc *BusClient) HandleIncomingMessage(subject string, data []byte) error {
	go func() {
		x_log.Info("Client", bc.ID, "received message for subject", subject)
	}()
	return nil
}

//-----------------------------------------
//  Utility Methods
//-----------------------------------------

// GetHandler retrieves the handler function for a subject.
func (bc *BusClient) GetHandler(subject string) (func(subject string, msg []byte), bool) {
	handler, exists := bc.msgHandlers[subject]
	return handler, exists
}

// GetClientID returns the client ID.
func (bc *BusClient) GetClientID() uint64 {
	return bc.ID
}

// ProcessBatch processes a message batch with subject-based handlers.
func (bc *BusClient) ProcessBatch(batch []typ.IRequest) error {
	for _, req := range batch {
		subject := req.GetSubject()
		data := req.GetData()
		handler, exists := bc.GetHandler(subject)
		if !exists {
			x_log.Error("Client", bc.ID, "No handler found for subject", subject)
			continue
		}
		handler(subject, data)
	}
	return nil
}

// SubscribeToTopic is an alias for Subscribe with added logging.
func (bc *BusClient) SubscribeToTopic(subject, queue string, handler func(subject string, msg []byte)) error {
	return bc.Subscribe(subject, queue, handler)
}
