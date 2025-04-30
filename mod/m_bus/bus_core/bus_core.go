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

type Bus struct {
	middlewareChain []typ.IMiddleware
	subscriptions   *bus_sub.Sublist
	secretKey       string
	messageChannel  chan typ.IRequest
	selfClient      typ.IBusClient
	clients         sync.Map
	batchBuffer     []typ.IRequest
	batchTicker     *time.Ticker
	mu              sync.Mutex
	cond            *sync.Cond
	maxGoroutines   int

	// Хост и Порт самой шины (сервер)
	Host string
	Port int

	// Сервер для прослушивания TCP-соединений
	listener net.Listener
}

// NewBus создает новый экземпляр шины с информацией о хосте и порте
// В функции NewBus
func NewBus(secretKey string, maxGoroutines int, host string, port int) *Bus {
	if !isValidIP(host) {
		x_log.Error("Invalid IP address. Please provide a valid IPv4 or IPv6 address.")
		return nil
	}

	// Создаем объект Bus
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

	// Инициализируем selfClient
	bus.InitClient()

	bus.cond = sync.NewCond(&bus.mu)

	go bus.ProcessMessages()

	x_log.Info("Bus created with secret key")

	return bus
}

// Инициализация клиента
func (b *Bus) InitClient() {
	if b.selfClient == nil {
		x_log.Info("Initializing self client...")
		b.selfClient = bus_client.NewBusClient(111, b, b.secretKey, 5, b.Host, b.Port)
	}
}

// Start запускает TCP-сервер на указанном хосте и порту
func (b *Bus) Start() error {
	// Форматируем адрес для сервера
	address := fmt.Sprintf("%s:%d", b.Host, b.Port)

	// Создаём новый TCP-сервер
	listener, err := net.Listen("tcp", address)
	if err != nil {
		x_log.Error(fmt.Sprintf("Error starting TCP server on %s: %v", address, err))
		return err
	}

	b.listener = listener

	// Запускаем горутину для принятия соединений
	go b.acceptConnections()

	x_log.Info(fmt.Sprintf("Bus server started on %s", address))
	return nil
}

// Stop останавливает TCP-сервер
func (b *Bus) Stop() error {
	// Останавливаем сервер
	if b.listener != nil {
		if err := b.listener.Close(); err != nil {
			x_log.Error(fmt.Sprintf("Error stopping server: %v", err))
			return err
		}
	}

	x_log.Info("Bus server stopped")
	return nil
}

// acceptConnections принимает входящие TCP-соединения
func (b *Bus) acceptConnections() {
	for {
		// Ожидаем новое подключение
		conn, err := b.listener.Accept()
		if err != nil {
			x_log.Error("Error accepting connection:", err)
			continue
		}

		// Обрабатываем новое подключение
		go b.handleConnection(conn)
	}
}

// handleConnection обрабатывает соединение с клиентом
func (b *Bus) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Читаем данные из соединения
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			x_log.Error("Error reading from connection:", err)
			return
		}

		// Получаем полученные данные
		data := buffer[:n]
		x_log.Info("Received data:", string(data))

		// Извлекаем тему и данные из полученного сообщения
		// Для примера, можно использовать первый байт как тему, остальные — как данные
		subject := fmt.Sprintf("topic_%s", string(data[:1])) // Пример: извлекаем тему из первого байта
		messageData := data[1:]                              // Пример: данные — это все, что после первого байта

		// Создаем запрос для дальнейшей обработки
		message := &bus_req.Request{
			Subject: subject,
			Data:    messageData,
		}

		// Обрабатываем сообщение
		err = b.ProcessMessage(message.Subject, message.Data)
		if err != nil {
			x_log.Error("Error processing message:", err)
			continue
		}

		// Публикуем сообщение для подписчиков
		err = b.Publish([]byte(subject), []byte("default_queue"), messageData)
		if err != nil {
			x_log.Error("Error publishing message:", err)
			continue
		}
	}
}

// isValidIP проверяет, является ли предоставленный хост валидным IP-адресом
func isValidIP(host string) bool {
	// Разбираем хост, чтобы проверить, является ли он валидным IP
	addr := net.ParseIP(host)
	return addr != nil
}

//-----------------------------------------
//  RemoveHandler - Removes a handler for a subject
//-----------------------------------------

// RemoveHandler removes a handler for a subject.
func (b *Bus) RemoveHandler(subject string) error {
	x_log.Info("No handler for subject", subject, "to remove")
	return nil
}

//-----------------------------------------
//  Use - Adds middleware to the bus
//-----------------------------------------

// Use adds middleware to the bus.
func (b *Bus) Use(middleware typ.IMiddleware) {
	b.middlewareChain = append(b.middlewareChain, middleware)
}

//-----------------------------------------
//  ProcessMessages - Processes messages asynchronously
//-----------------------------------------

// ProcessMessages processes messages from the channel asynchronously, with batching.
func (b *Bus) ProcessMessages() {
	x_log.Info("Bus is processing messages")

	// Limit the number of concurrent goroutines processing messages
	sem := make(chan struct{}, b.maxGoroutines) // Semaphore for controlling concurrency

	for {
		select {
		case req := <-b.messageChannel:
			sem <- struct{}{} // Acquire a slot
			go func(req typ.IRequest) {
				defer func() { <-sem }() // Release the slot
				b.batchBuffer = append(b.batchBuffer, req)

				// Process the batch if the buffer reaches a threshold (100 messages)
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
//  ProcessMessage - Processes a message and calls the handler
//-----------------------------------------

// ProcessMessage processes a message and calls the corresponding handler.
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
//  handleBatch - Processes a batch of messages
//-----------------------------------------

// handleBatch processes a batch of messages.
func (b *Bus) handleBatch(batch []typ.IRequest) {
	x_log.Info("Processing batch of", len(batch), "messages")
	for _, req := range batch {
		x_log.Info("Processing message:", string(req.GetData()))
	}
	x_log.Info("Finished processing batch")
}

// RetryPublish retries the message publishing attempts
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

// RetrySubscribe retries subscription attempts for a client
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

// Publish sends data to all subscribers
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

// retrySendToClient retries sending the message to the client with exponential backoff
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

// Subscribe adds a subscription for a client
func (b *Bus) Subscribe(subject []byte, queue []byte, client typ.IBusClient) error {
	subscription := typ.NewSubscription(subject, queue, client)

	if !isValidTopic(string(subject)) {
		return fmt.Errorf("invalid topic: %s", string(subject))
	}

	b.subscriptions.Insert(subscription)
	x_log.Info("Client successfully subscribed to subject", string(subject), "with queue", string(queue))
	return nil
}

// GetSubscriptions returns the list of subscriptions for a given subject
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

// Unsubscribe removes a subscription for a client
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

// GetMiddleware returns the middleware list
func (b *Bus) GetMiddleware() []typ.IMiddleware {
	return b.middlewareChain
}

// AddClient adds a client to the message bus
func (b *Bus) AddClient(client typ.IBusClient, id uint64) error {
	if _, exists := b.clients.Load(id); exists {
		return fmt.Errorf("client with ID %d already exists", id)
	}

	b.clients.Store(id, client)
	x_log.Info("Client with ID", id, "added to the bus")
	return nil
}

// RemoveClient removes a client from the message bus by its ID
func (b *Bus) RemoveClient(id uint64) error {
	if _, exists := b.clients.Load(id); exists {
		b.clients.Delete(id)
		x_log.Info("Client with ID", id, "removed from the bus")
		return nil
	}

	return fmt.Errorf("client with ID %d not found on the bus", id)
}

// GetClient retrieves a client by its ID
func (b *Bus) GetClient(id uint64) (typ.IBusClient, error) {
	client, exists := b.clients.Load(id)
	if !exists {
		return nil, fmt.Errorf("client with ID %d not found", id)
	}
	return client.(typ.IBusClient), nil
}

// GetClients returns all registered clients
func (b *Bus) GetClients() map[uint64]typ.IBusClient {
	clients := make(map[uint64]typ.IBusClient)
	b.clients.Range(func(key, value interface{}) bool {
		clients[key.(uint64)] = value.(typ.IBusClient)
		return true
	})
	return clients
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
