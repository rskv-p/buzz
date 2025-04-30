// file: buzz/typ/types.go

package typ

import (
	"context"
	"time"

	"github.com/rskv-p/buzz/pkg/x_log"
)

//-----------------------------------------
//  ActionHandler & IAction
//-----------------------------------------

// ActionHandler defines the function signature for action processing.
type ActionHandler func(IAction) any

// IAction defines the interface for application actions.
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

	// Input validation and parsing
	ValidateInputsNumber(length int) error
	NumberOfInputs() int
	NumberOfInputsIs(num int) bool
	InputNotNull(i int) error
	InputString(i int, defaults ...string) string
	InputInt(i int, defaults ...int) int
	InputUint32(i int, defaults ...uint32) uint32
	InputBool(i int, defaults ...bool) bool
	InputURL(i int, key string, defaults ...string) string
	InputMap(i int) map[string]any
	InputArray(i int) []any
	InputStrings(i int) []string
	InputsRecords(i int) []map[string]any
}

//-----------------------------------------
//  ClientFactory
//-----------------------------------------

// ClientFactory defines a function type for creating bus clients.
type ClientFactory func(id uint64, bus IBus) IBusClient

//-----------------------------------------
//  IBus
//-----------------------------------------

// IBus defines the interface for a message bus.
type IBus interface {
	Use(middleware IMiddleware)
	RemoveHandler(subject string) error
	ProcessMessage(subject string, msg []byte) error
	Subscribe(subject []byte, queue []byte, client IBusClient) error
	RetrySubscribe(subject []byte, queue []byte, retries int, delay time.Duration, client IBusClient) error
	Unsubscribe(subject []byte, queue []byte) error
	Publish(subject []byte, queue []byte, data []byte) error
	RetryPublish(subject []byte, queue []byte, data []byte, retries int, delay time.Duration) error
	GetSubscriptions(subject []byte) ([]*Subscription, error)
	GetMiddleware() []IMiddleware
	ProcessMessages()
	AddClient(client IBusClient, id uint64) error
	RemoveClient(id uint64) error
	GetClient(id uint64) (IBusClient, error)
	GetClients() map[uint64]IBusClient

	Start() error
	Stop() error
}

//-----------------------------------------
//  IMiddleware
//-----------------------------------------

// IMiddleware defines middleware for request processing.
type IMiddleware interface {
	Process(req IRequest) error
}

//-----------------------------------------
//  IBusClient
//-----------------------------------------

// IBusClient defines the interface for a message bus client.
type IBusClient interface {
	RegisterHandler(subject string, handler func(subject string, msg []byte)) error
	GetHandler(subject string) (func(subject string, msg []byte), bool)

	SubscribeToTopic(subject, queue string, handler func(subject string, msg []byte)) error
	RetrySubscribe(subject, queue string, retries int, delay time.Duration, handler func(subject string, msg []byte)) error

	PublishMessage(subject string, data []byte, queue string) error
	RetryPublish(subject string, data []byte, retries int, delay time.Duration, queue string) error

	HandleIncomingMessage(subject string, data []byte) error
	SendRequest(ctx context.Context, req IRequest) error
	Subscribe(subject, queue string, handler func(subject string, msg []byte)) error
	SendToMessageChannel(subject string, data []byte) error
	ProcessBatch(batch []IRequest) error
	GetClientID() uint64
}

//-----------------------------------------
//  IRequest
//-----------------------------------------

// IRequest defines the interface for working with a request.
type IRequest interface {
	RespondJSON(v any) error
	Error(code, description string, data []byte) error
	SetErrorHandler(f func(code, msg string, data any) error)
	SetHeader(key, value string) error
	Headers() (IRequestHeaders, error)
	GetSubject() string
	GetData() []byte
	GetReply() string
}

//-----------------------------------------
//  IRequestHeaders
//-----------------------------------------

// IRequestHeaders defines the interface for request headers.
type IRequestHeaders interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

//-----------------------------------------
//  Subscription
//-----------------------------------------

// Subscription represents a subject-queue-client tuple.
type Subscription struct {
	Subject []byte
	Queue   []byte
	Client  IBusClient
}

// NewSubscription creates a new subscription instance.
func NewSubscription(subject, queue []byte, client IBusClient) *Subscription {
	x_log.Info("Creating new subscription:", string(subject), string(queue))
	return &Subscription{Subject: subject, Queue: queue, Client: client}
}

//-----------------------------------------
//  IModule
//-----------------------------------------

// IModule defines the lifecycle of a functional module.
type IModule interface {
	GetName() string
	Init() error
	Start() error
	Stop() error

	RegisterAction(action IAction) error
	UnregisterAction(name string) error
	GetActions() []IAction
	GetAction(name string) (IAction, bool)
}

//-----------------------------------------
//  IService
//-----------------------------------------

// IService defines the lifecycle of a service.
type IService interface {
	GetName() string
	Start() error
	Stop()
}
