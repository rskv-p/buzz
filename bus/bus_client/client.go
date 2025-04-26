// file: buz/bus/bus_client/client.go

package bus_client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_req"
	"github.com/rskv-p/buzz/bus/bus_type"
)

//-----------------------------------------
//  BusClient Struct
//-----------------------------------------

// BusClient represents a client for interacting with the message bus.
type BusClient struct {
	ID                uint64                 // Client identifier
	Bus               bus_type.IBus          // Bus the client interacts with
	Subscriptions     *sync.Map              // Thread-safe subscriptions map (sync.Map)
	Client            bus_type.IBusClient    // Implementation of the bus client for subscription/publishing
	messageChannel    chan bus_type.IRequest // Channel for asynchronously processing messages
	secretKey         string                 // Secret key for authentication
	batchBuffer       []bus_type.IRequest    // Buffer for batch message processing
	batchTicker       *time.Ticker           // Timer for batch processing interval
	mu                sync.Mutex             // Mutex for synchronizing access to the client
	cond              *sync.Cond
	msgHandlers       map[string]func(subject string, msg []byte) // Handlers for each subject
	maxConcurrentMsgs int                                         // Max number of concurrent message handlers
}

//-----------------------------------------
//  BusClient Initialization
//-----------------------------------------

// NewBusClient creates a new instance of a bus client.
func NewBusClient(id uint64, bus bus_type.IBus, secretKey string, maxConcurrentMsgs int) bus_type.IBusClient {
	client := &BusClient{
		ID:                id,
		Bus:               bus,
		Subscriptions:     &sync.Map{},                       // Use sync.Map for subscriptions
		messageChannel:    make(chan bus_type.IRequest, 100), // Buffered channel for message processing
		secretKey:         secretKey,
		batchBuffer:       make([]bus_type.IRequest, 0),                      // Initialize batch buffer
		batchTicker:       time.NewTicker(5 * time.Second),                   // Timer for periodic batch processing
		msgHandlers:       make(map[string]func(subject string, msg []byte)), // Initialize handlers map
		maxConcurrentMsgs: maxConcurrentMsgs,
	}

	client.cond = sync.NewCond(&client.mu)

	// Start a goroutine to process messages
	go client.processMessages()

	bus_comm.Infof("BusClient %d created and ready to process messages", id)

	return client
}

//-----------------------------------------
//  Message Processing
//-----------------------------------------

// RegisterHandler registers a handler for a given subject on the client.
func (bc *BusClient) RegisterHandler(subject string, handler func(subject string, msg []byte)) error {
	if bc.msgHandlers == nil {
		bc.msgHandlers = make(map[string]func(subject string, msg []byte))
	}

	bc.msgHandlers[subject] = handler
	bus_comm.Infof("Handler registered for subject %s on client %d", subject, bc.ID)
	return nil
}

// processMessages processes received messages asynchronously and in batches.
func (bc *BusClient) processMessages() {
	bus_comm.Infof("Client %d started processing messages", bc.ID)

	// Create a semaphore to limit the number of concurrent goroutines
	sem := make(chan struct{}, bc.maxConcurrentMsgs)

	for {
		select {
		case req := <-bc.messageChannel:
			sem <- struct{}{} // Acquire a slot for the goroutine
			go func(req bus_type.IRequest) {
				defer func() { <-sem }() // Release the slot when done
				if request, ok := req.(*bus_req.Request); ok {
					bus_comm.Infof("Client %d received message for subject: %s", bc.ID, request.GetSubject())

					bc.batchBuffer = append(bc.batchBuffer, request)

					if len(bc.batchBuffer) >= 100 {
						bus_comm.Infof("Client %d batch buffer full, processing batch of %d messages", bc.ID, len(bc.batchBuffer))
						bc.handleBatch(bc.batchBuffer)
						bc.batchBuffer = nil
					}
				} else {
					bus_comm.Errorf("Client %d received invalid message type", bc.ID)
				}
			}(req)

		case <-bc.batchTicker.C:
			if len(bc.batchBuffer) > 0 {
				bus_comm.Infof("Client %d processing remaining messages in batch", bc.ID)
				bc.handleBatch(bc.batchBuffer)
				bc.batchBuffer = nil
			}
		}
	}
}

// handleBatch processes a batch of messages.
func (bc *BusClient) handleBatch(batch []bus_type.IRequest) {
	bus_comm.Infof("Client %d processing a batch of %d messages", bc.ID, len(batch))

	for _, req := range batch {
		bus_comm.Infof("Client %d processing message: %s", bc.ID, string(req.GetData()))

		subject := req.GetSubject()
		data := req.GetData()

		handler, exists := bc.GetHandler(subject)
		if !exists {
			bus_comm.Errorf("Client %d: No handler found for subject %s", bc.ID, subject)
			continue
		}

		handler(subject, data)
	}

	bus_comm.Infof("Client %d finished processing batch", bc.ID)
}

//-----------------------------------------
//  Subscription Management
//-----------------------------------------

// subscribeTopic subscribes the client to a specified subject with a given queue.
func (bc *BusClient) subscribeTopic(subject string, queue string) error {
	bus_comm.Infof("Client %d subscribing to subject %s with queue %s", bc.ID, subject, queue)

	// Use sync.Map to manage subscriptions efficiently
	subscriptions, _ := bc.Subscriptions.LoadOrStore(subject, []string{})
	existingQueues := subscriptions.([]string)

	for _, existingQueue := range existingQueues {
		if existingQueue == queue {
			bus_comm.Infof("Client %d is already subscribed to subject %s with queue %s", bc.ID, subject, queue)
			return nil
		}
	}

	// Add new queue to subscriptions
	bc.Subscriptions.Store(subject, append(existingQueues, queue))

	err := bc.Bus.Subscribe([]byte(subject), []byte(queue), bc)
	if err != nil {
		bus_comm.Errorf("Client %d: Subscription failed for subject %s with queue %s: %v", bc.ID, subject, queue, err)
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	bus_comm.Infof("Client %d successfully subscribed to subject %s with queue %s", bc.ID, subject, queue)
	return nil
}

// Subscribe subscribes the client to a specified subject with a given queue.
func (bc *BusClient) Subscribe(subject string, queue string) error {
	return bc.subscribeTopic(subject, queue)
}

// RetrySubscribe attempts to subscribe to a topic with retries.
func (bc *BusClient) RetrySubscribe(subject string, queue string, retries int, delay time.Duration) error {
	bus_comm.Infof("Client %d: Attempting to retry subscription to subject %s with queue %s", bc.ID, subject, queue)

	for i := 0; i < retries; i++ {
		err := bc.Subscribe(subject, queue)
		if err == nil {
			bus_comm.Infof("Client %d: Successfully subscribed to subject %s after %d attempts", bc.ID, subject, i+1)
			return nil
		}

		bus_comm.Errorf("Client %d: Subscription failed for %s, attempt %d: %v", bc.ID, subject, i+1, err)
		time.Sleep(delay)
	}

	return fmt.Errorf("client %d: failed to subscribe after %d attempts", bc.ID, retries)
}

//-----------------------------------------
//  Message Publishing
//-----------------------------------------

// PublishMessage publishes a message to the specified subject and queue.
func (bc *BusClient) PublishMessage(subject string, data []byte, queue string) error {
	bus_comm.Infof("Client %d publishing message to subject %s with queue %s", bc.ID, subject, queue)

	err := bc.Bus.Publish([]byte(subject), []byte(queue), data)
	if err != nil {
		bus_comm.Errorf("Client %d: Failed to publish message to subject %s with queue %s: %v", bc.ID, subject, queue, err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	bus_comm.Infof("Client %d successfully published message to subject %s with queue %s", bc.ID, subject, queue)
	return nil
}

// RetryPublish attempts to publish a message multiple times with retries.
func (bc *BusClient) RetryPublish(subject string, data []byte, retries int, delay time.Duration, queue string) error {
	currentDelay := delay

	bus_comm.Infof("Client %d: Attempting to retry publishing message to %s with queue %s", bc.ID, subject, queue)

	for i := 0; i < retries; i++ {
		err := bc.PublishMessage(subject, data, queue)
		if err == nil {
			bus_comm.Infof("Client %d: Successfully published message to %s after %d attempts", bc.ID, subject, i+1)
			return nil
		}

		bus_comm.Errorf("Client %d: Failed to publish to %s, attempt %d: %v", bc.ID, subject, i+1, err)

		if err.Error() == "channel full" {
			bus_comm.Warnf("Client %d: Message channel is full, retrying after delay: %v", bc.ID, currentDelay)
		}

		time.Sleep(currentDelay)
		currentDelay *= 2
	}

	return fmt.Errorf("client %d: failed to publish message after %d attempts", bc.ID, retries)
}

//-----------------------------------------
//  Incoming Message Handling
//-----------------------------------------

// HandleIncomingMessage asynchronously handles messages for a given subject.
func (bc *BusClient) HandleIncomingMessage(subject string, data []byte) error {
	bus_comm.Infof("Client %d: Handling incoming message for subject %s", bc.ID, subject)

	go func() {
		bus_comm.Infof("Client %d received message for subject %s: %s", bc.ID, subject, string(data))
	}()

	return nil
}

//-----------------------------------------
//  Request Sending
//-----------------------------------------

// SendRequest sends a request through the bus and processes it through the middleware chain.
func (bc *BusClient) SendRequest(ctx context.Context, req bus_type.IRequest) error {
	for _, middleware := range bc.Bus.GetMiddleware() {
		select {
		case <-ctx.Done():
			bus_comm.Errorf("Client %d: Context cancelled: %v", bc.ID, ctx.Err())
			return ctx.Err()
		default:
			if err := middleware.Process(req); err != nil {
				bus_comm.Errorf("Client %d: Error processing request through middleware: %v", bc.ID, err)
				return err
			}
		}
	}

	bus_comm.Infof("Client %d: Request processed successfully", bc.ID)
	return nil
}

// SendToMessageChannel sends data to the client's message channel.
func (bc *BusClient) SendToMessageChannel(subject string, data []byte) error {
	bus_comm.Infof("Client %d: Sending message to message channel", bc.ID)

	select {
	case bc.messageChannel <- &bus_req.Request{
		Subject: subject,
		Data:    data,
	}:
		bus_comm.Infof("Client %d: Message sent to message channel successfully", bc.ID)
		return nil
	default:
		bus_comm.Errorf("Client %d: Failed to send message to message channel: channel is full", bc.ID)
		return fmt.Errorf("message channel is full")
	}
}

//-----------------------------------------
//  IBusClient Interface Implementation
//-----------------------------------------

// SubscribeToTopic subscribes the client to a specified subject with a queue.
func (bc *BusClient) SubscribeToTopic(subject string, queue string) error {
	return bc.Subscribe(subject, queue)
}

// ProcessBatch processes a batch of messages asynchronously.
func (bc *BusClient) ProcessBatch(batch []bus_type.IRequest) error {
	bus_comm.Infof("Client %d is processing a batch of %d messages", bc.ID, len(batch))

	for _, req := range batch {
		bus_comm.Infof("Processing message: %s", string(req.GetData()))

		subject := req.GetSubject()
		data := req.GetData()

		handler, exists := bc.GetHandler(subject)
		if !exists {
			bus_comm.Errorf("No handler found for subject %s", subject)
			continue
		}

		handler(subject, data)
	}

	bus_comm.Infof("Finished processing batch")
	return nil
}

// GetClientID retrieves the client ID.
func (bc *BusClient) GetClientID() uint64 {
	return bc.ID
}

// GetHandler retrieves the handler for a specific subject in the client
func (bc *BusClient) GetHandler(subject string) (func(subject string, msg []byte), bool) {
	handler, exists := bc.msgHandlers[subject]
	return handler, exists
}
