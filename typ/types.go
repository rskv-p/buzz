// file: buzz/typ/types.go

package typ

import (
	"context"
	"time"

	"github.com/rskv-p/buzz/pkg/x_log"
)

//---------------------
// Action and Handler
//---------------------

// Handler defines the function signature for handlers that process actions.
type ActionHandler func(IAction) any

// IAction defines the core behavior that an action should implement.
type IAction interface {
	IsPublic() bool
	GetInputs() []any
	SetInputs(inputs []any) error
	Run() func(IAction) any
	GetMethod() string
	GetName() string
	GetModuleName() string
	WithContext(ctx context.Context) IAction
	Context() context.Context
	GetOutput() any
	Dispose()
	//Inputs
	ValidateInputsNumber(length int) error                 // ValidateInputsNumber checks if the number of inputs is sufficient
	NumberOfInputs() int                                   // NumberOfInputs returns the number of inputs to the action
	NumberOfInputsIs(num int) bool                         // NumberOfInputsIs checks if the number of inputs matches the specified number
	InputNotNull(i int) error                              // InputNotNull ensures that a specific input is not nil
	InputString(i int, defaults ...string) string          // InputString parses a string input at the specified index
	InputInt(i int, defaults ...int) int                   // InputInt parses an integer input at the specified index
	InputUint32(i int, defaults ...uint32) uint32          // InputUint32 parses a uint32 input at the specified index
	InputBool(i int, defaults ...bool) bool                // InputBool parses a boolean input at the specified index
	InputURL(i int, key string, defaults ...string) string // InputURL parses a URL input at the specified index and retrieves a specific key
	InputMap(i int) map[string]any                         // InputMap parses an input as a map
	InputArray(i int) []any                                // InputArray parses an input as an array
	InputStrings(i int) []string                           // InputStrings parses an input as an array of strings
	InputsRecords(i int) []map[string]any                  // InputsRecords parses an input as an array of records (maps)
}

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
	Use(middleware IMiddleware)                                                                             // Use adds middleware to the bus.
	RemoveHandler(subject string) error                                                                     // RemoveHandler removes a handler for a specific message subject.
	ProcessMessage(subject string, msg []byte) error                                                        // ProcessMessage processes a message and invokes the appropriate handler.
	Subscribe(subject []byte, queue []byte, client IBusClient) error                                        // Subscribe subscribes a client to a subject with a queue.
	RetrySubscribe(subject []byte, queue []byte, retries int, delay time.Duration, client IBusClient) error // RetrySubscribe retries subscription attempts for a client.
	Unsubscribe(subject []byte, queue []byte) error                                                         // Unsubscribe removes a subscription for a client.
	Publish(subject []byte, queue []byte, data []byte) error                                                // Publish publishes raw data to all clients subscribed to the subject and queue.
	RetryPublish(subject []byte, queue []byte, data []byte, retries int, delay time.Duration) error         // RetryPublish retries the publish attempt.
	GetSubscriptions(subject []byte) ([]*Subscription, error)                                               // GetSubscriptions returns the list of subscriptions for a specific subject.
	GetMiddleware() []IMiddleware                                                                           // GetMiddleware returns all middleware added to the bus.
	ProcessMessages()                                                                                       // ProcessMessages processes messages asynchronously via a channel.
	AddClient(client IBusClient, id uint64) error                                                           // AddClient adds a client to the bus.
	RemoveClient(id uint64) error                                                                           // RemoveClient removes a client from the bus.
	GetClient(id uint64) (IBusClient, error)                                                                // GetClient retrieves a client by its ID.
	GetClients() map[uint64]IBusClient                                                                      // GetClients returns all registered clients.
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
	SubscribeToTopic(string, string, func(string, []byte)) error                                                                   // SubscribeToTopic subscribes the client to a specific topic with a queue.
	RetrySubscribe(subject string, queue string, retries int, delay time.Duration, handler func(subject string, msg []byte)) error // RetrySubscribe retries the subscription to a topic with a queue multiple times.
	PublishMessage(subject string, data []byte, queue string) error                                                                // PublishMessage publishes a message to the specified topic.
	RetryPublish(subject string, data []byte, retries int, delay time.Duration, queue string) error                                // RetryPublish retries the publishing of a message multiple times.
	HandleIncomingMessage(subject string, data []byte) error                                                                       // HandleIncomingMessage handles incoming messages asynchronously on the given topic.
	SendRequest(ctx context.Context, req IRequest) error                                                                           // SendRequest sends a request through the bus and processes it via middleware.
	Subscribe(subject string, queue string, handler func(subject string, msg []byte)) error                                        // Subscribe subscribes the client to a specified topic with a queue.
	SendToMessageChannel(string, []byte) error                                                                                     // SendToMessageChannel sends data to the client's message channel.
	ProcessBatch(batch []IRequest) error                                                                                           // ProcessBatch processes a batch of messages asynchronously.
	GetClientID() uint64                                                                                                           // GetClientID retrieves the client ID.
}

//-------------------------------------------------
// IRequest - Interface for working with requests
//-------------------------------------------------

// IRequest defines the interface for working with requests.
type IRequest interface {
	RespondJSON(v any) error                                  // RespondJSON sends a JSON response.
	Error(code, description string, data []byte) error        // Error sends an error in JSON format.
	SetErrorHandler(f func(code, msg string, data any) error) // SetErrorHandler sets an error handler.
	SetHeader(key, value string) error                        // SetHeader sets a header for the request.
	Headers() (IRequestHeaders, error)                        // Headers returns the headers of the request.
	GetSubject() string                                       // GetSubject retrieves the subject of the request.
	GetData() []byte                                          // GetData retrieves the data of the request.
	GetReply() string                                         // GetReply retrieves the reply for the request.
}

//-------------------------------------------------
// IRequestHeaders - Interface for working with headers
//-------------------------------------------------

// IRequestHeaders defines the interface for handling request headers.
type IRequestHeaders interface {
	Get(key string) (string, error) // Get retrieves the value of a header by its key.
	Set(key, value string) error    // Set sets a header for the request.
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

// NewSubscription creates a new subscription with the provided subject, queue, and client.
func NewSubscription(subject, queue []byte, client IBusClient) *Subscription {
	x_log.Info("Creating new subscription:", string(subject), string(queue))

	return &Subscription{
		Subject: subject,
		Queue:   queue,
		Client:  client,
	}
}

//---------------------
// Interfaces
//---------------------

// IModule defines the lifecycle of a module, including initialization, starting, stopping, and action management.
type IModule interface {
	GetName() string // Returns the name of the module
	Init() error     // Initializes the module
	Start() error    // Starts the module (sets status to running)
	Stop() error     // Stops the module (sets status to down)

	// Actions Management
	RegisterAction(action IAction) error   // Registers a new action
	UnregisterAction(name string) error    // Unregisters an action by name
	GetActions() []IAction                 // Returns all registered actions
	GetAction(name string) (IAction, bool) // Retrieves a specific action by name
}

// IService defines a common interface for services
type IService interface {
	GetName() string // Return service name
	Start() error    // Start the service
	Stop()           // Stop the service
}
