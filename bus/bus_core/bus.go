// file: buz/bus/bus_core/bus.go

package bus_core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rskv-p/buzz/bus/bus_client"
	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_sub"
	"github.com/rskv-p/buzz/bus/bus_type"
)

//-----------------------------------------
//  Bus - Represents the message bus and its core operations
//-----------------------------------------

// Bus represents the message bus and its core operations, including middleware, subscriptions, and client management.
type Bus struct {
	middlewareChain []bus_type.IMiddleware // Chain of middleware functions
	subscriptions   *bus_sub.Sublist       // Subscription list for topics and queues
	secretKey       string                 // Secret key for the bus
	messageChannel  chan bus_type.IRequest // Channel for asynchronously processing messages
	selfClient      bus_type.IBusClient    // The main bus client
	clients         sync.Map               // Map of clients
	batchBuffer     []bus_type.IRequest    // Buffer for batch message processing
	batchTicker     *time.Ticker           // Timer for batch processing interval
	mu              sync.Mutex             // Mutex for synchronizing access to the bus
	cond            *sync.Cond             // Condition variable for synchronization
	maxGoroutines   int                    // Max concurrent goroutines for message processing
}

//-----------------------------------------
//  NewBus - Creates a new bus instance
//-----------------------------------------

// NewBus creates a new instance of the bus.
func NewBus(secretKey string, selfClient bus_type.IBusClient, maxGoroutines int) *Bus {
	bus := &Bus{
		middlewareChain: make([]bus_type.IMiddleware, 0),
		subscriptions:   bus_sub.NewSublist(100),
		secretKey:       secretKey,
		messageChannel:  make(chan bus_type.IRequest, 100), // Increase the size of the message channel
		clients:         sync.Map{},
		selfClient:      selfClient,
		batchBuffer:     make([]bus_type.IRequest, 0),
		batchTicker:     time.NewTicker(5 * time.Second),
		maxGoroutines:   maxGoroutines, // Limit the number of concurrent goroutines
	}

	bus.cond = sync.NewCond(&bus.mu)

	go bus.ProcessMessages()

	return bus
}

//-----------------------------------------
//  RetryPublish - Retries message publishing
//-----------------------------------------

// RetryPublish retries the message publishing attempts with exponential backoff.
func (b *Bus) RetryPublish(subject []byte, queue []byte, data []byte, retries int, delay time.Duration) error {
	// Validate the topic (subject)
	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	bus_comm.Infof("Attempting to publish message to subject %s with queue %s, initial delay: %v", string(subject), string(queue), delay)

	for i := 0; i < retries; i++ {
		err := b.Publish(subject, queue, data)
		if err == nil {
			bus_comm.Infof("Message successfully published to subject %s with queue %s after %d attempt(s)", string(subject), string(queue), i+1)
			return nil
		}

		bus_comm.Errorf("Publishing error for subject %s, attempt %d: %v", string(subject), i+1, err)
		bus_comm.Infof("Retrying to publish message after delay: %v", delay)

		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf("failed to publish message to subject %s with queue %s after %d attempts", string(subject), string(queue), retries)
}

//-----------------------------------------
//  RetrySubscribe - Retries subscription attempts
//-----------------------------------------

// RetrySubscribe retries subscription attempts for a client.
func (b *Bus) RetrySubscribe(subject []byte, queue []byte, retries int, delay time.Duration, client bus_type.IBusClient) error {
	// Validate the topic (subject)
	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	bus_comm.Infof("Attempting to subscribe client to subject %s with queue %s, initial delay: %v", string(subject), string(queue), delay)

	for i := 0; i < retries; i++ {
		err := b.Subscribe(subject, queue, client)
		if err == nil {
			bus_comm.Infof("Client successfully subscribed to subject %s with queue %s after %d attempt(s)", string(subject), string(queue), i+1)
			return nil
		}

		bus_comm.Errorf("Subscription attempt %d failed for subject %s with queue %s: %v", i+1, string(subject), string(queue), err)
		bus_comm.Infof("Retrying subscription after delay: %v", delay)

		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf("failed to subscribe to subject %s with queue %s after %d attempts", string(subject), string(queue), retries)
}

//-----------------------------------------
//  Publish - Publishes message to subscribers
//-----------------------------------------

// Publish sends data to all subscribers subscribed to the subject and queue.
func (b *Bus) Publish(subject []byte, queue []byte, data []byte) error {
	// Validate the topic (subject)
	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	bus_comm.Infof("Publishing message to subject %s with queue %s", string(subject), string(queue))

	sublistResult := b.subscriptions.MatchWithQueue(subject, queue)
	if len(sublistResult.Psubs) == 0 {
		bus_comm.Warnf("No subscriptions found for subject %s with queue %s", string(subject), string(queue))
	}

	for _, sub := range sublistResult.Psubs {
		client := sub.Client
		bus_comm.Debugf("Sending data to client %v", client)

		err := client.(*bus_client.BusClient).SendToMessageChannel(string(subject), data)
		if err != nil {
			bus_comm.Errorf("Failed to send message to client %v: %v", client, err)
			if retryErr := b.retrySendToClient(client, data, subject, queue); retryErr != nil {
				bus_comm.Errorf("Failed to send message after retries: %v", retryErr)
			}
		}
	}

	return nil
}

//-----------------------------------------
//  retrySendToClient - Retries sending message to client
//-----------------------------------------

// retrySendToClient retries sending the message to the client with exponential backoff
func (b *Bus) retrySendToClient(client bus_type.IBusClient, data []byte, subject []byte, queue []byte) error {
	const maxRetries = 5
	delay := 100 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		err := client.(*bus_client.BusClient).SendToMessageChannel(string(subject), data)
		if err == nil {
			bus_comm.Infof("Message successfully sent to client after %d retries", i+1)
			return nil
		}
		bus_comm.Errorf("Retry %d failed to send message to client: %v", i+1, err)
		time.Sleep(delay)
		delay *= 2
	}
	return fmt.Errorf("failed to send message to client after %d retries", maxRetries)
}

//-----------------------------------------
//  ProcessMessages - Processes messages asynchronously
//-----------------------------------------

// ProcessMessages processes messages from the channel asynchronously, with batching.
func (b *Bus) ProcessMessages() {
	bus_comm.Infof("Bus is processing messages")

	// Limit the number of concurrent goroutines processing messages
	sem := make(chan struct{}, b.maxGoroutines) // Semaphore for controlling concurrency

	for {
		select {
		case req := <-b.messageChannel:
			sem <- struct{}{} // Acquire a slot
			go func(req bus_type.IRequest) {
				defer func() { <-sem }() // Release the slot
				b.batchBuffer = append(b.batchBuffer, req)

				// Process the batch if the buffer reaches a threshold (100 messages)
				if len(b.batchBuffer) >= 100 {
					bus_comm.Infof("Batch size reached, processing %d messages", len(b.batchBuffer))
					b.handleBatch(b.batchBuffer)
					b.batchBuffer = nil
				}
			}(req)

		case <-b.batchTicker.C:
			if len(b.batchBuffer) > 0 {
				bus_comm.Infof("Periodic batch processing with %d messages", len(b.batchBuffer))
				b.handleBatch(b.batchBuffer)
				b.batchBuffer = nil
			}
		}
	}
}

//-----------------------------------------
//  handleBatch - Processes a batch of messages
//-----------------------------------------

// handleBatch processes a batch of messages.
func (b *Bus) handleBatch(batch []bus_type.IRequest) {
	bus_comm.Infof("Processing batch of %d messages", len(batch))
	for _, req := range batch {
		bus_comm.Infof("Processing message: %s", string(req.GetData()))
	}
	bus_comm.Infof("Finished processing batch")
}

//-----------------------------------------
//  isValidTopic - Validates the topic
//-----------------------------------------

// isValidTopic checks if the topic is valid.
func isValidTopic(topic string) bool {
	// For this example, invalid topics will be those starting with "invalid."
	if strings.HasPrefix(topic, "invalid.") {
		return false
	}
	return true
}

//-----------------------------------------
//  GetMiddleware - Returns the middleware list
//-----------------------------------------

// GetMiddleware returns the list of all middleware added to the bus.
func (b *Bus) GetMiddleware() []bus_type.IMiddleware {
	return b.middlewareChain
}

//-----------------------------------------
//  AddClient - Adds a client to the bus
//-----------------------------------------

// AddClient adds a client to the message bus.
func (b *Bus) AddClient(client bus_type.IBusClient, id uint64) error {
	if _, exists := b.clients.Load(id); exists {
		return fmt.Errorf("client with ID %d already exists", id)
	}

	b.clients.Store(id, client)
	bus_comm.Infof("Client with ID %d added to the bus", id)
	return nil
}

//-----------------------------------------
//  RemoveClient - Removes a client from the bus
//-----------------------------------------

// RemoveClient removes a client from the message bus by its ID.
func (b *Bus) RemoveClient(id uint64) error {
	if _, exists := b.clients.Load(id); exists {
		b.clients.Delete(id)
		bus_comm.Infof("Client with ID %d removed from the bus", id)
		return nil
	}

	return fmt.Errorf("client with ID %d not found on the bus", id)
}

//-----------------------------------------
//  GetClient - Retrieves a client by its ID
//-----------------------------------------

// GetClient retrieves a client by its ID.
func (b *Bus) GetClient(id uint64) (bus_type.IBusClient, error) {
	client, exists := b.clients.Load(id)
	if !exists {
		return nil, fmt.Errorf("client with ID %d not found", id)
	}
	return client.(bus_type.IBusClient), nil
}

//-----------------------------------------
//  GetClients - Returns all registered clients
//-----------------------------------------

// GetClients returns all registered clients.
func (b *Bus) GetClients() map[uint64]bus_type.IBusClient {
	clients := make(map[uint64]bus_type.IBusClient)
	b.clients.Range(func(key, value interface{}) bool {
		clients[key.(uint64)] = value.(bus_type.IBusClient)
		return true
	})
	return clients
}

//-----------------------------------------
//  Use - Adds middleware to the bus
//-----------------------------------------

// Use adds middleware to the bus.
func (b *Bus) Use(middleware bus_type.IMiddleware) {
	b.middlewareChain = append(b.middlewareChain, middleware)
}

//-----------------------------------------
//  RemoveHandler - Removes a handler for a subject
//-----------------------------------------

// RemoveHandler removes a handler for a subject.
func (b *Bus) RemoveHandler(subject string) error {
	bus_comm.Infof("No handler for subject %s to remove", subject)
	return nil
}

//-----------------------------------------
//  ProcessMessage - Processes a message and calls the handler
//-----------------------------------------

// ProcessMessage processes a message and calls the corresponding handler.
func (b *Bus) ProcessMessage(subject string, msg []byte) error {
	bus_comm.Infof("Processing message for subject %s", subject)

	b.clients.Range(func(key, value interface{}) bool {
		client := value.(bus_type.IBusClient)

		if handler, exists := client.GetHandler(subject); exists {
			bus_comm.Infof("Found handler for subject %s, invoking handler", subject)
			handler(subject, msg)
		} else {
			bus_comm.Warnf("No handler found for client %v for subject %s", key, subject)
		}

		return true
	})

	return nil
}

//-----------------------------------------
//  Subscribe - Adds a subscription for a client
//-----------------------------------------

// Subscribe adds a subscription for a client to a specific topic with a queue.
func (b *Bus) Subscribe(subject []byte, queue []byte, client bus_type.IBusClient) error {
	subscription := bus_type.NewSubscription(subject, queue, client)

	// Validate the topic (subject)
	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	b.subscriptions.Insert(subscription)
	bus_comm.Infof("Client successfully subscribed to subject %s with queue %s", string(subject), string(queue))
	return nil
}

//-----------------------------------------
//  GetSubscriptions - Returns subscriptions for a subject
//-----------------------------------------

// GetSubscriptions returns the list of subscriptions for a given subject.
func (b *Bus) GetSubscriptions(subject []byte) ([]*bus_type.Subscription, error) {
	bus_comm.Infof("Fetching subscriptions for subject %s", string(subject))

	sublistResult := b.subscriptions.Match(subject)
	if len(sublistResult.Psubs) == 0 {
		bus_comm.Warnf("No subscriptions found for subject %s", string(subject))
		return nil, fmt.Errorf("no subscriptions found for subject %s", string(subject))
	}

	bus_comm.Infof("Found %d subscriptions for subject %s", len(sublistResult.Psubs), string(subject))
	return sublistResult.Psubs, nil
}

//-----------------------------------------
//  Unsubscribe - Removes a subscription for a client
//-----------------------------------------

// Unsubscribe removes a subscription for a client from a specific topic with a queue.
func (b *Bus) Unsubscribe(subject []byte, queue []byte) error {
	subscription := &bus_type.Subscription{Subject: subject, Queue: queue}

	err := b.subscriptions.Remove(subscription)
	if err != nil {
		bus_comm.Errorf("Failed to remove subscription for subject %s with queue %s: %v", string(subject), string(queue), err)
		return fmt.Errorf("failed to remove subscription: %w", err)
	}

	bus_comm.Infof("Successfully unsubscribed from subject %s with queue %s", string(subject), string(queue))
	return nil
}
