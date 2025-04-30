// file: buzz/mod/m_bus/bus_core/bus_core.go

package bus_core

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/rskv-p/buzz/mod/m_bus/bus_client"
	"github.com/rskv-p/buzz/mod/m_bus/bus_req"
	"github.com/rskv-p/buzz/mod/m_bus/bus_sub"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

var _ typ.IBus = (*Bus)(nil)

//-----------------------------------------
//  Bus Struct
//-----------------------------------------

type Bus struct {
	middlewareChain []typ.IMiddleware
	subscriptions   *bus_sub.Sublist
	secretKey       string
	messageChannel  chan typ.IRequest
	clients         sync.Map
	batchBuffer     []typ.IRequest
	batchTicker     *time.Ticker
	mu              sync.Mutex
	cond            *sync.Cond
	maxGoroutines   int
	Host            string
	Port            int
	listener        net.Listener
}

//-----------------------------------------
//  NewBus
//-----------------------------------------

// NewBus creates a new Bus instance with specified host and port.
func NewBus(secretKey string, maxGoroutines int, host string, port int) *Bus {
	if !isValidIP(host) {
		x_log.Error("Invalid IP address. Please provide a valid IPv4 or IPv6 address.")
		return nil
	}

	bus := &Bus{
		middlewareChain: make([]typ.IMiddleware, 0),
		subscriptions:   bus_sub.NewSublist(100),
		secretKey:       secretKey,
		messageChannel:  make(chan typ.IRequest, 100),
		clients:         sync.Map{},
		batchBuffer:     make([]typ.IRequest, 0),
		batchTicker:     time.NewTicker(5 * time.Second),
		maxGoroutines:   maxGoroutines,
		Host:            host,
		Port:            port,
	}

	bus.cond = sync.NewCond(&bus.mu)

	go bus.ProcessMessages()

	x_log.Info("Bus created with secret key")

	return bus
}

//-----------------------------------------
//  Start
//-----------------------------------------

// Start launches the TCP server on the specified host and port.
func (b *Bus) Start() error {
	var address string
	if strings.Contains(b.Host, ":") {
		address = fmt.Sprintf("[%s]:%d", b.Host, b.Port)
	} else {
		address = fmt.Sprintf("%s:%d", b.Host, b.Port)
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		x_log.Error(fmt.Sprintf("Error starting TCP server on %s: %v", address, err))
		return err
	}

	b.listener = listener

	go b.acceptConnections()

	x_log.Info(fmt.Sprintf("Bus server started on %s", address))
	return nil
}

//-----------------------------------------
//  Stop
//-----------------------------------------

// Stop terminates the TCP server.
func (b *Bus) Stop() error {
	if b.listener != nil {
		if err := b.listener.Close(); err != nil {
			x_log.Error(fmt.Sprintf("Error stopping server: %v", err))
			return err
		}
	}

	x_log.Info("Bus server stopped")
	return nil
}

//-----------------------------------------
//  acceptConnections
//-----------------------------------------

// acceptConnections handles incoming TCP connections.
func (b *Bus) acceptConnections() {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			x_log.Error("Error accepting connection:", err)
			continue
		}

		go b.handleConnection(conn)
	}
}

//-----------------------------------------
//  handleConnection
//-----------------------------------------

// handleConnection reads and processes data from client connection.
func (b *Bus) handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			x_log.Error("Error reading from connection:", err)
			return
		}

		data := buffer[:n]
		x_log.Info("Received data:", string(data))

		subject := fmt.Sprintf("topic_%s", string(data[:1]))
		messageData := data[1:]

		message := &bus_req.Request{
			Subject: subject,
			Data:    messageData,
		}

		err = b.ProcessMessage(message.Subject, message.Data)
		if err != nil {
			x_log.Error("Error processing message:", err)
			continue
		}

		err = b.Publish([]byte(subject), []byte("default_queue"), messageData)
		if err != nil {
			x_log.Error("Error publishing message:", err)
			continue
		}
	}
}

//-----------------------------------------
//  isValidIP
//-----------------------------------------

// isValidIP checks whether the given host is a valid IP.
func isValidIP(host string) bool {
	addr := net.ParseIP(host)
	return addr != nil
}

//-----------------------------------------
//  RemoveHandler
//-----------------------------------------

// RemoveHandler removes a handler for the subject (stub).
func (b *Bus) RemoveHandler(subject string) error {
	x_log.Info("No handler for subject", subject, "to remove")
	return nil
}

//-----------------------------------------
//  Use
//-----------------------------------------

// Use adds a middleware to the bus.
func (b *Bus) Use(middleware typ.IMiddleware) {
	b.middlewareChain = append(b.middlewareChain, middleware)
}

//-----------------------------------------
//  ProcessMessages
//-----------------------------------------

// ProcessMessages processes messages asynchronously with batching.
func (b *Bus) ProcessMessages() {
	x_log.Info("Bus is processing messages")

	sem := make(chan struct{}, b.maxGoroutines)

	for {
		select {
		case req := <-b.messageChannel:
			sem <- struct{}{}
			go func(req typ.IRequest) {
				defer func() { <-sem }()
				b.batchBuffer = append(b.batchBuffer, req)

				if len(b.batchBuffer) >= 100 {
					x_log.Info("Batch size reached, processing", len(b.batchBuffer), "messages")
					b.handleBatch(b.batchBuffer)
					b.batchBuffer = nil
				}
			}(req)

		case <-b.batchTicker.C:
			if len(b.batchBuffer) > 0 {
				x_log.Info("Periodic batch processing with", len(b.batchBuffer), "messages")
				b.handleBatch(b.batchBuffer)
				b.batchBuffer = nil
			}
		}
	}
}

//-----------------------------------------
//  ProcessMessage
//-----------------------------------------

// ProcessMessage routes a message to the corresponding handler.
func (b *Bus) ProcessMessage(subject string, msg []byte) error {
	x_log.Info("Processing message for subject", subject)

	b.clients.Range(func(key, value interface{}) bool {
		client := value.(typ.IBusClient)

		if handler, exists := client.GetHandler(subject); exists {
			x_log.Info("Found handler for subject", subject, "invoking handler")
			handler(subject, msg)
		} else {
			x_log.Warn("No handler found for client", key, "for subject", subject)
		}

		return true
	})

	return nil
}

//-----------------------------------------
//  handleBatch
//-----------------------------------------

// handleBatch processes a batch of messages.
func (b *Bus) handleBatch(batch []typ.IRequest) {
	x_log.Info("Processing batch of", len(batch), "messages")
	for _, req := range batch {
		x_log.Info("Processing message:", string(req.GetData()))
	}
	x_log.Info("Finished processing batch")
}

//-----------------------------------------
//  AddClient
//-----------------------------------------

// AddClient adds a client to the message bus.
func (b *Bus) AddClient(client typ.IBusClient, id uint64) error {
	if _, exists := b.clients.Load(id); exists {
		return fmt.Errorf("client with ID %d already exists", id)
	}

	b.clients.Store(id, client)
	x_log.Info("Client with ID", id, "added to the bus")
	return nil
}

//-----------------------------------------
//  Publish
//-----------------------------------------

// Publish sends data to all subscribers of the given subject and queue.
func (b *Bus) Publish(subject []byte, queue []byte, data []byte) error {
	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	x_log.Info("Publishing message to subject", string(subject), "with queue", string(queue))

	sublistResult := b.subscriptions.MatchWithQueue(subject, queue)
	if len(sublistResult.Psubs) == 0 {
		x_log.Warn("No subscriptions found for subject", string(subject), "with queue", string(queue))
	}

	for _, sub := range sublistResult.Psubs {
		client := sub.Client
		x_log.Debug("Sending data to client", client)

		err := client.(*bus_client.BusClient).SendToMessageChannel(string(subject), data)
		if err != nil {
			x_log.Error("Failed to send message to client", client, ":", err)
			if retryErr := b.retrySendToClient(client, data, subject, queue); retryErr != nil {
				x_log.Error("Failed to send message after retries:", retryErr)
			}
		}
	}

	return nil
}

//-----------------------------------------
//  isValidTopic
//-----------------------------------------

// isValidTopic checks if the topic name is valid.
func isValidTopic(topic string) bool {
	// Reject topics starting with "invalid."
	return !strings.HasPrefix(topic, "invalid.")
}

//-----------------------------------------
//  retrySendToClient
//-----------------------------------------

// retrySendToClient attempts to resend the message to the client with exponential backoff.
func (b *Bus) retrySendToClient(client typ.IBusClient, data []byte, subject []byte, queue []byte) error {
	const maxRetries = 5
	delay := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		err := client.(*bus_client.BusClient).SendToMessageChannel(string(subject), data)
		if err == nil {
			x_log.Info("Message successfully sent to client after", i+1, "retries")
			return nil
		}
		x_log.Error("Retry", i+1, "failed to send message to client:", err)
		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf("failed to send message to client after %d retries", maxRetries)
}

//-----------------------------------------
//  GetClient
//-----------------------------------------

// GetClient retrieves a client by its ID.
func (b *Bus) GetClient(id uint64) (typ.IBusClient, error) {
	client, exists := b.clients.Load(id)
	if !exists {
		return nil, fmt.Errorf("client with ID %d not found", id)
	}
	return client.(typ.IBusClient), nil
}

//-----------------------------------------
//  GetClients
//-----------------------------------------

// GetClients returns all registered clients as a map.
func (b *Bus) GetClients() map[uint64]typ.IBusClient {
	clients := make(map[uint64]typ.IBusClient)
	b.clients.Range(func(key, value interface{}) bool {
		clients[key.(uint64)] = value.(typ.IBusClient)
		return true
	})
	return clients
}

//-----------------------------------------
//  GetMiddleware
//-----------------------------------------

// GetMiddleware returns the list of middleware attached to the bus.
func (b *Bus) GetMiddleware() []typ.IMiddleware {
	return b.middlewareChain
}

//-----------------------------------------
//  GetSubscriptions
//-----------------------------------------

// GetSubscriptions returns all subscriptions for the given subject.
func (b *Bus) GetSubscriptions(subject []byte) ([]*typ.Subscription, error) {
	x_log.Info("Fetching subscriptions for subject", string(subject))

	sublistResult := b.subscriptions.Match(subject)
	if len(sublistResult.Psubs) == 0 {
		x_log.Warn("No subscriptions found for subject", string(subject))
		return nil, fmt.Errorf("no subscriptions found for subject %s", string(subject))
	}

	x_log.Info("Found", len(sublistResult.Psubs), "subscriptions for subject", string(subject))
	return sublistResult.Psubs, nil
}

//-----------------------------------------
//  RemoveClient
//-----------------------------------------

// RemoveClient removes a client from the bus by its ID.
func (b *Bus) RemoveClient(id uint64) error {
	if _, exists := b.clients.Load(id); exists {
		b.clients.Delete(id)
		x_log.Info("Client with ID", id, "removed from the bus")
		return nil
	}

	return fmt.Errorf("client with ID %d not found on the bus", id)
}

//-----------------------------------------
//  RetryPublish
//-----------------------------------------

// RetryPublish attempts to publish a message with retries and exponential backoff.
func (b *Bus) RetryPublish(subject []byte, queue []byte, data []byte, retries int, delay time.Duration) error {
	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	x_log.Info("Attempting to publish message to subject", string(subject), "with queue", string(queue), "initial delay:", delay)

	for i := 0; i < retries; i++ {
		err := b.Publish(subject, queue, data)
		if err == nil {
			x_log.Info("Message successfully published to subject", string(subject), "with queue", string(queue), "after", i+1, "attempt(s)")
			return nil
		}

		x_log.Error("Publishing error for subject", string(subject), "attempt", i+1, ":", err)
		x_log.Info("Retrying to publish message after delay:", delay)

		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf("failed to publish message to subject %s with queue %s after %d attempts", string(subject), string(queue), retries)
}

//-----------------------------------------
//  RetrySubscribe
//-----------------------------------------

// RetrySubscribe attempts to subscribe a client to a subject and queue with retries.
func (b *Bus) RetrySubscribe(subject []byte, queue []byte, retries int, delay time.Duration, client typ.IBusClient) error {
	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	x_log.Info("Attempting to subscribe client to subject", string(subject), "with queue", string(queue), "initial delay:", delay)

	for i := 0; i < retries; i++ {
		err := b.Subscribe(subject, queue, client)
		if err == nil {
			x_log.Info("Client successfully subscribed to subject", string(subject), "with queue", string(queue), "after", i+1, "attempt(s)")
			return nil
		}

		x_log.Error("Subscription attempt", i+1, "failed for subject", string(subject), "with queue", string(queue), ":", err)
		x_log.Info("Retrying subscription after delay:", delay)

		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf("failed to subscribe to subject %s with queue %s after %d attempts", string(subject), string(queue), retries)
}

//-----------------------------------------
//  Subscribe
//-----------------------------------------

// Subscribe adds a new subscription for the given subject and queue.
func (b *Bus) Subscribe(subject []byte, queue []byte, client typ.IBusClient) error {
	subscription := typ.NewSubscription(subject, queue, client)

	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	b.subscriptions.Insert(subscription)
	x_log.Info("Client successfully subscribed to subject", string(subject), "with queue", string(queue))
	return nil
}

//-----------------------------------------
//  Unsubscribe
//-----------------------------------------

// Unsubscribe removes a client's subscription for the given subject and queue.
func (b *Bus) Unsubscribe(subject []byte, queue []byte) error {
	subscription := &typ.Subscription{Subject: subject, Queue: queue}

	err := b.subscriptions.Remove(subscription)
	if err != nil {
		x_log.Error("Failed to remove subscription for subject", string(subject), "with queue", string(queue), ":", err)
		return fmt.Errorf("failed to remove subscription: %w", err)
	}

	x_log.Info("Successfully unsubscribed from subject", string(subject), "with queue", string(queue))
	return nil
}
