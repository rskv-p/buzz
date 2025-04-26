// file: buz/bus/bus_type/types.go

package bus_type

import (
	"context"
	"time"

	"github.com/rskv-p/buzz/bus/bus_comm"
)

//-------------------------------------------------
// ClientFactory - Type definition for client creation
//-------------------------------------------------

// ClientFactory defines a function type for creating new clients.
type ClientFactory func(id uint64, bus IBus) IBusClient

//-------------------------------------------------
// IBus - Interface for message bus
//-------------------------------------------------

// IBus defines the interface for interacting with the message bus.
type IBus interface {
	// Use adds middleware to the bus.
	Use(middleware IMiddleware)

	// RemoveHandler removes a handler for a specific message subject.
	RemoveHandler(subject string) error

	// ProcessMessage processes a message and invokes the appropriate handler.
	ProcessMessage(subject string, msg []byte) error

	// Subscribe subscribes a client to a subject with a queue.
	Subscribe(subject []byte, queue []byte, client IBusClient) error

	// RetrySubscribe retries subscription attempts for a client.
	RetrySubscribe(subject []byte, queue []byte, retries int, delay time.Duration, client IBusClient) error

	// Unsubscribe removes a subscription for a client.
	Unsubscribe(subject []byte, queue []byte) error

	// Publish publishes raw data to all clients subscribed to the subject and queue.
	Publish(subject []byte, queue []byte, data []byte) error

	// RetryPublish retries the publish attempt.
	RetryPublish(subject []byte, queue []byte, data []byte, retries int, delay time.Duration) error

	// GetSubscriptions returns the list of subscriptions for a specific subject.
	GetSubscriptions(subject []byte) ([]*Subscription, error)

	// GetMiddleware returns all middleware added to the bus.
	GetMiddleware() []IMiddleware

	// ProcessMessages processes messages asynchronously via a channel.
	ProcessMessages()

	// AddClient adds a client to the bus.
	AddClient(client IBusClient, id uint64) error

	RemoveClient(id uint64) error

	// GetClient retrieves a client by its ID.
	GetClient(id uint64) (IBusClient, error)

	// GetClients returns all registered clients.
	GetClients() map[uint64]IBusClient

	// ProcessBatch processes a batch of messages.
	//ProcessBatch(batch []IRequest) error
}

//-------------------------------------------------
// IMiddleware - Interface for middleware
//-------------------------------------------------

// IMiddleware defines the interface for middleware.
type IMiddleware interface {
	// Process processes the request through the middleware.
	Process(req IRequest) error
}

//-------------------------------------------------
// IBusClient - Interface for bus client
//-------------------------------------------------

// IBusClient defines the interface for interacting with a message bus client.
type IBusClient interface {
	RegisterHandler(subject string, handler func(subject string, msg []byte)) error

	GetHandler(subject string) (func(subject string, msg []byte), bool)
	// SubscribeToTopic subscribes the client to a specific topic with a queue.
	SubscribeToTopic(subject string, queue string) error

	// RetrySubscribe retries the subscription to a topic with a queue multiple times.
	RetrySubscribe(subject string, queue string, retries int, delay time.Duration) error

	// PublishMessage publishes a message to the specified topic.
	PublishMessage(subject string, data []byte, queue string) error

	// RetryPublish retries the publishing of a message multiple times.
	RetryPublish(subject string, data []byte, retries int, delay time.Duration, queue string) error

	// HandleIncomingMessage handles incoming messages asynchronously on the given topic.
	HandleIncomingMessage(subject string, data []byte) error

	// SendRequest sends a request through the bus and processes it via middleware.
	SendRequest(ctx context.Context, req IRequest) error

	// Subscribe subscribes the client to a specified topic with a queue.
	Subscribe(subject string, queue string) error

	// SendToMessageChannel sends data to the client's message channel.
	SendToMessageChannel(string, []byte) error

	// ProcessBatch processes a batch of messages asynchronously.
	ProcessBatch(batch []IRequest) error

	// GetClientID retrieves the client ID.
	GetClientID() uint64
}

//-------------------------------------------------
// IRequest - Interface for working with requests
//-------------------------------------------------

// IRequest defines the interface for working with requests.
type IRequest interface {
	// RespondJSON sends a JSON response.
	RespondJSON(v any) error

	// Error sends an error in JSON format.
	Error(code, description string, data []byte) error

	// SetErrorHandler sets an error handler.
	SetErrorHandler(f func(code, msg string, data any) error)

	// SetHeader sets a header for the request.
	SetHeader(key, value string) error

	// Headers returns the headers of the request.
	Headers() (IRequestHeaders, error)

	// GetSubject retrieves the subject of the request.
	GetSubject() string

	// GetData retrieves the data of the request.
	GetData() []byte

	// GetReply retrieves the reply for the request.
	GetReply() string
}

//-------------------------------------------------
// IRequestHeaders - Interface for working with headers
//-------------------------------------------------

// IRequestHeaders defines the interface for handling request headers.
type IRequestHeaders interface {
	// Get retrieves the value of a header by its key.
	Get(key string) (string, error)

	// Set sets a header for the request.
	Set(key, value string) error
}

//-------------------------------------------------
// Subscription - Represents a subscription to a topic
//-------------------------------------------------

// Subscription represents a subscription to a subject with an associated client and queue.
type Subscription struct {
	Subject []byte     // The subject to which the client is subscribed
	Queue   []byte     // The queue associated with the subscription
	Client  IBusClient // The client associated with the subscription
}

//-------------------------------------------------
// NewSubscription - Creates a new subscription
//-------------------------------------------------

// NewSubscription creates a new subscription with the provided subject, queue, and client.
func NewSubscription(subject, queue []byte, client IBusClient) *Subscription {
	bus_comm.Infof("Creating new subscription: subject = %s, queue = %s", string(subject), string(queue))

	return &Subscription{
		Subject: subject,
		Queue:   queue,
		Client:  client,
	}
}
