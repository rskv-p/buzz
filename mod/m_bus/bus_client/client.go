// file:buzz/mod/m_bus/bus_client/client.go

package bus_client

import (
	"context"
	"fmt"
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

// BusClient represents a client for interacting with the message bus.
type BusClient struct {
	ID                uint64            // Client identifier
	Bus               typ.IBus          // Bus the client interacts with
	Subscriptions     *sync.Map         // Thread-safe subscriptions map (sync.Map)
	Client            typ.IBusClient    // Implementation of the bus client for subscription/publishing
	messageChannel    chan typ.IRequest // Channel for asynchronously processing messages
	secretKey         string            // Secret key for authentication
	batchBuffer       []typ.IRequest    // Buffer for batch message processing
	batchTicker       *time.Ticker      // Timer for batch processing interval
	mu                sync.Mutex        // Mutex for synchronizing access to the client
	cond              *sync.Cond
	msgHandlers       map[string]func(subject string, msg []byte) // Handlers for each subject
	maxConcurrentMsgs int                                         // Max number of concurrent message handlers
}

//-----------------------------------------
//  BusClient Initialization
//-----------------------------------------

// NewBusClient creates a new instance of a bus client.
func NewBusClient(id uint64, bus typ.IBus, secretKey string, maxConcurrentMsgs int) typ.IBusClient {
	client := &BusClient{
		ID:                id,
		Bus:               bus,
		Subscriptions:     &sync.Map{},                  // Use sync.Map for subscriptions
		messageChannel:    make(chan typ.IRequest, 100), // Buffered channel for message processing
		secretKey:         secretKey,
		batchBuffer:       make([]typ.IRequest, 0),                           // Initialize batch buffer
		batchTicker:       time.NewTicker(5 * time.Second),                   // Timer for periodic batch processing
		msgHandlers:       make(map[string]func(subject string, msg []byte)), // Initialize handlers map
		maxConcurrentMsgs: maxConcurrentMsgs,
	}

	client.cond = sync.NewCond(&client.mu)

	// Start a goroutine to process messages
	go client.processMessages()

	x_log.Info("BusClient", id, "created and ready to process messages")

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
	x_log.Info("Handler registered for subject", subject, "on client", bc.ID)
	return nil
}

// processMessages processes received messages asynchronously and in batches.
func (bc *BusClient) processMessages() {
	x_log.Info("Client", bc.ID, "started processing messages")

	// Create a semaphore to limit the number of concurrent goroutines
	sem := make(chan struct{}, bc.maxConcurrentMsgs)

	for {
		select {
		case req := <-bc.messageChannel:
			sem <- struct{}{} // Acquire a slot for the goroutine
			go func(req typ.IRequest) {
				defer func() { <-sem }() // Release the slot when done
				if request, ok := req.(*bus_req.Request); ok {
					x_log.Info("Client", bc.ID, "received message for subject:", request.GetSubject())

					bc.batchBuffer = append(bc.batchBuffer, request)

					if len(bc.batchBuffer) >= 100 {
						x_log.Info("Client", bc.ID, "batch buffer full, processing batch of", len(bc.batchBuffer), "messages")
						bc.handleBatch(bc.batchBuffer)
						bc.batchBuffer = nil
					}
				} else {
					x_log.Error("Client", bc.ID, "received invalid message type")
				}
			}(req)

		case <-bc.batchTicker.C:
			if len(bc.batchBuffer) > 0 {
				x_log.Info("Client", bc.ID, "processing remaining messages in batch")
				bc.handleBatch(bc.batchBuffer)
				bc.batchBuffer = nil
			}
		}
	}
}

// handleBatch processes a batch of messages.
func (bc *BusClient) handleBatch(batch []typ.IRequest) {
	x_log.Info("Client", bc.ID, "processing a batch of", len(batch), "messages")

	for _, req := range batch {
		x_log.Info("Client", bc.ID, "processing message:", string(req.GetData()))

		subject := req.GetSubject()
		data := req.GetData()

		handler, exists := bc.GetHandler(subject)
		if !exists {
			x_log.Error("Client", bc.ID, "No handler found for subject", subject)
			continue
		}

		handler(subject, data)
	}

	x_log.Info("Client", bc.ID, "finished processing batch")
}

//-----------------------------------------
//  Subscription Management
//-----------------------------------------

// subscribeTopic subscribes the client to a specified subject with a given queue.
func (bc *BusClient) subscribeTopic(subject string, queue string, handler func(subject string, msg []byte)) error {
	x_log.Info("Client", bc.ID, "subscribing to subject", subject, "with queue", queue)

	// Use sync.Map to manage subscriptions efficiently
	subscriptions, _ := bc.Subscriptions.LoadOrStore(subject, []string{})
	existingQueues := subscriptions.([]string)

	// Check if the client is already subscribed to this subject and queue
	for _, existingQueue := range existingQueues {
		if existingQueue == queue {
			x_log.Info("Client", bc.ID, "is already subscribed to subject", subject, "with queue", queue)
			return nil
		}
	}

	// Add new queue to subscriptions
	bc.Subscriptions.Store(subject, append(existingQueues, queue))

	// Register the handler for this subject
	err := bc.RegisterHandler(subject, handler)
	if err != nil {
		x_log.Error("Client", bc.ID, "Failed to register handler for subject", subject, ":", err)
		return fmt.Errorf("failed to register handler: %w", err)
	}

	// Subscribe the client to the subject with the queue
	err = bc.Bus.Subscribe([]byte(subject), []byte(queue), bc)
	if err != nil {
		x_log.Error("Client", bc.ID, "Subscription failed for subject", subject, "with queue", queue, ":", err)
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	x_log.Info("Client", bc.ID, "successfully subscribed to subject", subject, "with queue", queue)
	return nil
}

// Subscribe subscribes the client to a specified subject with a given queue and handler.
func (bc *BusClient) Subscribe(subject string, queue string, handler func(subject string, msg []byte)) error {
	return bc.subscribeTopic(subject, queue, handler)
}

// RetrySubscribe attempts to subscribe to a topic with retries.
func (bc *BusClient) RetrySubscribe(subject string, queue string, retries int, delay time.Duration, handler func(subject string, msg []byte)) error {
	x_log.Info("Client", bc.ID, "Attempting to retry subscription to subject", subject, "with queue", queue)

	for i := 0; i < retries; i++ {
		// Attempt to subscribe with the handler
		err := bc.Subscribe(subject, queue, handler)
		if err == nil {
			x_log.Info("Client", bc.ID, "Successfully subscribed to subject", subject, "after", i+1, "attempts")
			return nil
		}

		x_log.Error("Client", bc.ID, "Subscription failed for", subject, "attempt", i+1, ":", err)
		time.Sleep(delay)
	}

	return fmt.Errorf("client %d: failed to subscribe after %d attempts", bc.ID, retries)
}

//-----------------------------------------
//  Message Publishing
//-----------------------------------------

// PublishMessage publishes a message to the specified subject and queue.
func (bc *BusClient) PublishMessage(subject string, data []byte, queue string) error {
	x_log.Info("Client", bc.ID, "publishing message to subject", subject, "with queue", queue)

	err := bc.Bus.Publish([]byte(subject), []byte(queue), data)
	if err != nil {
		x_log.Error("Client", bc.ID, "Failed to publish message to subject", subject, "with queue", queue, ":", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	x_log.Info("Client", bc.ID, "successfully published message to subject", subject, "with queue", queue)
	return nil
}

// RetryPublish attempts to publish a message multiple times with retries.
func (bc *BusClient) RetryPublish(subject string, data []byte, retries int, delay time.Duration, queue string) error {
	currentDelay := delay

	x_log.Info("Client", bc.ID, "Attempting to retry publishing message to", subject, "with queue", queue)

	for i := 0; i < retries; i++ {
		err := bc.PublishMessage(subject, data, queue)
		if err == nil {
			x_log.Info("Client", bc.ID, "Successfully published message to", subject, "after", i+1, "attempts")
			return nil
		}

		x_log.Error("Client", bc.ID, "Failed to publish to", subject, "attempt", i+1, ":", err)

		if err.Error() == "channel full" {
			x_log.Warn("Client", bc.ID, "Message channel is full, retrying after delay:", currentDelay)
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
	x_log.Info("Client", bc.ID, "Handling incoming message for subject", subject)

	go func() {
		x_log.Info("Client", bc.ID, "received message for subject", subject, ":", string(data))
	}()

	return nil
}

//-----------------------------------------
//  Request Sending
//-----------------------------------------

// SendRequest sends a request through the bus and processes it through the middleware chain.
func (bc *BusClient) SendRequest(ctx context.Context, req typ.IRequest) error {
	for _, middleware := range bc.Bus.GetMiddleware() {
		select {
		case <-ctx.Done():
			x_log.Error("Client", bc.ID, "Context cancelled:", ctx.Err())
			return ctx.Err()
		default:
			if err := middleware.Process(req); err != nil {
				x_log.Error("Client", bc.ID, "Error processing request through middleware:", err)
				return err
			}
		}
	}

	x_log.Info("Client", bc.ID, "Request processed successfully")
	return nil
}

// SendToMessageChannel sends data to the client's message channel.
func (bc *BusClient) SendToMessageChannel(subject string, data []byte) error {
	x_log.Info("Client", bc.ID, "Sending message to message channel")

	select {
	case bc.messageChannel <- &bus_req.Request{
		Subject: subject,
		Data:    data,
	}:
		x_log.Info("Client", bc.ID, "Message sent to message channel successfully")
		return nil
	default:
		x_log.Error("Client", bc.ID, "Failed to send message to message channel: channel is full")
		return fmt.Errorf("message channel is full")
	}
}

//-----------------------------------------
//  IBusClient Interface Implementation
//-----------------------------------------

// SubscribeToTopic subscribes the client to a specified subject with a queue and registers a handler.
func (bc *BusClient) SubscribeToTopic(subject string, queue string, handler func(subject string, msg []byte)) error {
	return bc.Subscribe(subject, queue, handler)
}

// ProcessBatch processes a batch of messages asynchronously.
func (bc *BusClient) ProcessBatch(batch []typ.IRequest) error {
	x_log.Info("Client", bc.ID, "is processing a batch of", len(batch), "messages")

	for _, req := range batch {
		x_log.Info("Processing message:", string(req.GetData()))

		subject := req.GetSubject()
		data := req.GetData()

		handler, exists := bc.GetHandler(subject)
		if !exists {
			x_log.Error("No handler found for subject", subject)
			continue
		}

		handler(subject, data)
	}

	x_log.Info("Finished processing batch")
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
