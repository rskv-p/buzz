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

// BusClient представляет клиента для взаимодействия с шиной сообщений.
type BusClient struct {
	ID                uint64            // Идентификатор клиента
	Bus               typ.IBus          // Шина, с которой взаимодействует клиент
	Subscriptions     *sync.Map         // Потокобезопасная карта подписок (sync.Map)
	Client            typ.IBusClient    // Реализация клиента шины для подписок/публикации
	messageChannel    chan typ.IRequest // Канал для асинхронной обработки сообщений
	secretKey         string            // Секретный ключ для аутентификации
	batchBuffer       []typ.IRequest    // Буфер для обработки сообщений в пакетах
	batchTicker       *time.Ticker      // Таймер для периодической обработки пакетов
	mu                sync.Mutex        // Мьютекс для синхронизации доступа к клиенту
	cond              *sync.Cond
	msgHandlers       map[string]func(subject string, msg []byte) // Обработчики для каждого типа сообщения
	maxConcurrentMsgs int                                         // Максимальное количество параллельных обработчиков сообщений
	// Новое поле для хранения сетевого соединения
	Connection net.Conn // Соединение с шиной
}

//-----------------------------------------
//  Инициализация BusClient
//-----------------------------------------

// NewBusClient создает новый экземпляр клиента шины.
func NewBusClient(id uint64, bus typ.IBus, secretKey string, maxConcurrentMsgs int, host string, port int) typ.IBusClient {
	// Форматируем адрес для соединения (IPv4 или IPv6)
	address := formatAddress(host, port)

	// Устанавливаем соединение через net.Dial
	conn, err := net.Dial("tcp", address)
	if err != nil {
		x_log.Error("Error connecting to bus:", err)
		return nil
	}

	// Инициализируем BusClient
	client := &BusClient{
		ID:                id,
		Bus:               bus,
		Subscriptions:     &sync.Map{},                  // Используем sync.Map для подписок
		messageChannel:    make(chan typ.IRequest, 100), // Буферизированный канал для обработки сообщений
		secretKey:         secretKey,
		batchBuffer:       make([]typ.IRequest, 0),                           // Инициализируем буфер для пакетов
		batchTicker:       time.NewTicker(5 * time.Second),                   // Таймер для периодической обработки пакетов
		msgHandlers:       make(map[string]func(subject string, msg []byte)), // Инициализируем карту обработчиков
		maxConcurrentMsgs: maxConcurrentMsgs,
		Connection:        conn, // Сохраняем соединение
	}

	client.cond = sync.NewCond(&client.mu)

	// Запускаем горутину для обработки сообщений
	go client.processMessages()

	x_log.Info("BusClient", id, "created and ready to process messages")

	return client
}

// formatAddress форматирует адрес для IPv4 или IPv6
func formatAddress(host string, port int) string {
	// Проверяем, является ли хост IPv6-адресом
	if strings.Contains(host, ":") {
		// Для IPv6 адреса необходимо обернуть его в квадратные скобки
		return fmt.Sprintf("[%s]:%d", host, port)
	}

	// Для IPv4-адреса просто возвращаем host:port
	return fmt.Sprintf("%s:%d", host, port)
}

//-----------------------------------------
//  Обработка сообщений
//-----------------------------------------

// RegisterHandler регистрирует обработчик для определенной темы на клиенте.
func (bc *BusClient) RegisterHandler(subject string, handler func(subject string, msg []byte)) error {
	if bc.msgHandlers == nil {
		bc.msgHandlers = make(map[string]func(subject string, msg []byte))
	}

	bc.msgHandlers[subject] = handler
	x_log.Info("Handler registered for subject", subject, "on client", bc.ID)
	return nil
}

// processMessages обрабатывает полученные сообщения асинхронно и пакетами.
func (bc *BusClient) processMessages() {
	x_log.Info("Client", bc.ID, "started processing messages")

	// Создаем семафор для ограничения количества параллельных горутин
	sem := make(chan struct{}, bc.maxConcurrentMsgs)

	for {
		select {
		case req := <-bc.messageChannel:
			sem <- struct{}{} // Захватываем слот для горутины
			go func(req typ.IRequest) {
				defer func() { <-sem }() // Освобождаем слот, когда горутина завершена
				if request, ok := req.(*bus_req.Request); ok {
					x_log.Info("Client", bc.ID, "received message for subject:", request.GetSubject())

					bc.batchBuffer = append(bc.batchBuffer, request)

					// Если буфер набрал 100 сообщений, обрабатываем пакет
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

// handleBatch обрабатывает пакет сообщений.
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
//  Управление подписками
//-----------------------------------------

// subscribeTopic подписывает клиента на определенную тему с заданной очередью.
func (bc *BusClient) subscribeTopic(subject string, queue string, handler func(subject string, msg []byte)) error {
	x_log.Info("Client", bc.ID, "subscribing to subject", subject, "with queue", queue)

	// Используем sync.Map для эффективного управления подписками
	subscriptions, _ := bc.Subscriptions.LoadOrStore(subject, []string{})
	existingQueues := subscriptions.([]string)

	// Проверяем, подписан ли клиент уже на эту тему и очередь
	for _, existingQueue := range existingQueues {
		if existingQueue == queue {
			x_log.Info("Client", bc.ID, "is already subscribed to subject", subject, "with queue", queue)
			return nil
		}
	}

	// Добавляем новую очередь в подписки для данной темы
	bc.Subscriptions.Store(subject, append(existingQueues, queue))

	// Регистрируем обработчик для этой темы
	err := bc.RegisterHandler(subject, handler)
	if err != nil {
		x_log.Error("Client", bc.ID, "Failed to register handler for subject", subject, ":", err)
		return fmt.Errorf("failed to register handler: %w", err)
	}

	// Подписываем клиента на тему с очередью
	err = bc.Bus.Subscribe([]byte(subject), []byte(queue), bc)
	if err != nil {
		x_log.Error("Client", bc.ID, "Subscription failed for subject", subject, "with queue", queue, ":", err)
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	x_log.Info("Client", bc.ID, "successfully subscribed to subject", subject, "with queue", queue)
	return nil
}

// Subscribe подписывает клиента на указанную тему с очередью и обработчиком.
func (bc *BusClient) Subscribe(subject string, queue string, handler func(subject string, msg []byte)) error {
	return bc.subscribeTopic(subject, queue, handler)
}

// RetrySubscribe пытается подписаться на тему с повторными попытками.
func (bc *BusClient) RetrySubscribe(subject string, queue string, retries int, delay time.Duration, handler func(subject string, msg []byte)) error {
	x_log.Info("Client", bc.ID, "Attempting to retry subscription to subject", subject, "with queue", queue)

	for i := 0; i < retries; i++ {
		// Попытка подписки с обработчиком
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
//  Публикация сообщений
//-----------------------------------------

// PublishMessage публикует сообщение на указанную тему и очередь.
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

// RetryPublish пытается несколько раз опубликовать сообщение с повторными попытками.
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
//  Обработка входящих сообщений
//-----------------------------------------

// HandleIncomingMessage асинхронно обрабатывает сообщения для указанной темы.
func (bc *BusClient) HandleIncomingMessage(subject string, data []byte) error {
	x_log.Info("Client", bc.ID, "Handling incoming message for subject", subject)

	go func() {
		x_log.Info("Client", bc.ID, "received message for subject", subject, ":", string(data))
	}()

	return nil
}

//-----------------------------------------
//  Отправка запросов
//-----------------------------------------

// SendRequest отправляет запрос через шину и обрабатывает его через цепочку промежуточного ПО.
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

// SendToMessageChannel отправляет данные в канал сообщений клиента.
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

// GetClientID получает идентификатор клиента.
func (bc *BusClient) GetClientID() uint64 {
	return bc.ID
}

// GetHandler получает обработчик для конкретной темы в клиенте.
func (bc *BusClient) GetHandler(subject string) (func(subject string, msg []byte), bool) {
	handler, exists := bc.msgHandlers[subject]
	return handler, exists
}

// ProcessBatch асинхронно обрабатывает пакет сообщений, извлекая обработчики для каждой темы.
func (bc *BusClient) ProcessBatch(batch []typ.IRequest) error {
	x_log.Info("Client", bc.ID, "is processing a batch of", len(batch), "messages")

	for _, req := range batch {
		x_log.Info("Client", bc.ID, "processing message:", string(req.GetData()))

		subject := req.GetSubject() // Извлекаем тему из сообщения
		data := req.GetData()       // Извлекаем данные сообщения

		// Получаем обработчик для текущей темы
		handler, exists := bc.GetHandler(subject)
		if !exists {
			// Если обработчик не найден, выводим ошибку и продолжаем
			x_log.Error("Client", bc.ID, "No handler found for subject", subject)
			continue
		}

		// Вызываем обработчик для обработки данных
		handler(subject, data)
	}

	x_log.Info("Client", bc.ID, "finished processing batch")
	return nil
}

// SubscribeToTopic подписывает клиента на заданную тему с конкретной очередью и регистрирует обработчик.
func (bc *BusClient) SubscribeToTopic(subject string, queue string, handler func(subject string, msg []byte)) error {
	x_log.Info("Client", bc.ID, "subscribing to subject", subject, "with queue", queue)

	// Используем sync.Map для эффективного управления подписками
	subscriptions, _ := bc.Subscriptions.LoadOrStore(subject, []string{})
	existingQueues := subscriptions.([]string)

	// Проверяем, если клиент уже подписан на эту тему и очередь
	for _, existingQueue := range existingQueues {
		if existingQueue == queue {
			// Если подписка уже существует, выходим
			x_log.Info("Client", bc.ID, "is already subscribed to subject", subject, "with queue", queue)
			return nil
		}
	}

	// Добавляем новую очередь в список подписок для данной темы
	bc.Subscriptions.Store(subject, append(existingQueues, queue))

	// Регистрируем обработчик для этой темы
	err := bc.RegisterHandler(subject, handler)
	if err != nil {
		// Если регистрация обработчика не удалась, выводим ошибку
		x_log.Error("Client", bc.ID, "Failed to register handler for subject", subject, ":", err)
		return fmt.Errorf("failed to register handler: %w", err)
	}

	// Подписываем клиента на тему с указанной очередью
	err = bc.Bus.Subscribe([]byte(subject), []byte(queue), bc)
	if err != nil {
		// Если подписка на шину не удалась, выводим ошибку
		x_log.Error("Client", bc.ID, "Subscription failed for subject", subject, "with queue", queue, ":", err)
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Логируем успешную подписку
	x_log.Info("Client", bc.ID, "successfully subscribed to subject", subject, "with queue", queue)
	return nil
}
