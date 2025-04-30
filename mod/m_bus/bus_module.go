// file: buzz/mod/m_bus/bus_module.go

package m_bus

import (
	"fmt"
	"net/http"

	"github.com/rskv-p/buzz/act" // Assuming act.Action is imported from this package
	"github.com/rskv-p/buzz/mod"
	"github.com/rskv-p/buzz/mod/m_bus/bus_client"
	"github.com/rskv-p/buzz/mod/m_bus/bus_core"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/pkg/x_str"
	"github.com/rskv-p/buzz/typ"
)

var _ typ.IModule = (*BusModule)(nil) // Ensure RequestHeaders implements IRequestHeaders

//---------------------
// Error Handling
//---------------------

// ErrInvalidActionType is returned when an action cannot be cast to *act.Action.
var ErrInvalidActionType = fmt.Errorf("invalid action type")

// BusModule combines Module and Bus together
type BusModule struct {
	mod.Module               // <-- now an embedded Module
	Bus        *bus_core.Bus // Embedded bus (message processing)
	busClient  typ.IBusClient
}

// NewBusModule creates and returns a new BusModule with all actions registered.
func NewBusModule(name string, secretKey string, maxGoroutines int, host string, port int) *BusModule {
	// Create bus with host and port for both IPv4 and IPv6
	bus := bus_core.NewBus(secretKey, maxGoroutines, host, port)

	busClient := bus_client.NewBusClient(1, bus, secretKey, 10, host, port) // ID = 1, can use another ID

	// Build topic names for GET, POST, DELETE, PUT requests
	topicNameGet, err := x_str.BuildTopicName(name, "demo")
	if err != nil {
		x_log.Error("Failed to build topic name for GET:", err)
	}
	topicNamePost, err := x_str.BuildTopicName(name, "demo")
	if err != nil {
		x_log.Error("Failed to build topic name for POST:", err)
	}
	topicNameDelete, err := x_str.BuildTopicName(name, "demo")
	if err != nil {
		x_log.Error("Failed to build topic name for DELETE:", err)
	}
	topicNamePut, err := x_str.BuildTopicName(name, "demo")
	if err != nil {
		x_log.Error("Failed to build topic name for PUT:", err)
	}

	actions := map[string]typ.IAction{
		topicNameGet: &act.Action{
			Name:   topicNameGet,
			Public: true,
			Func:   HandleGetDemo,
			Method: http.MethodGet,
		},
		topicNamePost: &act.Action{
			Name:   topicNamePost,
			Public: true,
			Func:   HandleSetDemo,
			Method: http.MethodPost,
		},
		topicNameDelete: &act.Action{
			Name:   topicNameDelete,
			Public: true,
			Func:   HandleDeleteDemo,
			Method: http.MethodDelete,
		},
		topicNamePut: &act.Action{
			Name:   topicNamePut,
			Public: true,
			Func:   HandleUpdateDemo,
			Method: http.MethodPut,
		},
	}

	module := &mod.Module{
		Name:    name,
		Actions: actions,
	}

	return &BusModule{
		Module:    *module,
		Bus:       bus,
		busClient: busClient,
	}
}

//---------------------
// Helper Functions
//---------------------

// HandleGetDemo processes the GET request to fetch some demo data.
func HandleGetDemo(action typ.IAction) any {
	x_log.Info("Handling GET request for bus.demo.get")
	// Just a mock response for demo purposes
	return map[string]interface{}{
		"status":  "success",
		"message": "GET request handled",
	}
}

// HandleSetDemo processes the POST request to set some demo data.
func HandleSetDemo(action typ.IAction) any {
	x_log.Info("Handling POST request for bus.demo.set")

	// Just a mock response for demo purposes
	return map[string]interface{}{
		"status":  "success",
		"message": "POST request handled",
	}
}

// HandleDeleteDemo processes the DELETE request for demo data.
func HandleDeleteDemo(action typ.IAction) any {
	x_log.Info("Handling DELETE request for bus.demo.delete")

	// Just a mock response for demo purposes
	return map[string]interface{}{
		"status":  "success",
		"message": "DELETE request handled",
	}
}

// HandleUpdateDemo processes the PUT request to update demo data.
func HandleUpdateDemo(action typ.IAction) any {
	x_log.Info("Handling PUT request for bus.demo.update")

	// Just a mock response for demo purposes
	return map[string]interface{}{
		"status":  "success",
		"message": "PUT request handled",
	}
}

// HandleAction performs the action and returns the result.
func HandleAction(action typ.IAction) (any, error) {
	if action == nil {
		return nil, fmt.Errorf("action is nil")
	}

	return action.Run()(action), nil
}

// RegisterPublicActions registers actions that are public and available via REST API.

func (bm *BusModule) RegisterActionsOnBus() {
	for _, action := range bm.Module.GetActions() {

		// Get the base name of the action
		baseSubject := action.GetName()
		// Add the method to the topic name
		subjectWithMethod, err := x_str.BuildTopicName(baseSubject, action.GetMethod())
		if err != nil {
			x_log.Error("Failed to build full bus subject for action:", baseSubject, "Error:", err)
			continue
		}

		handler := func(subject string, msg []byte) {
			x_log.Info("Executing action for subject:", subject)

			// You can parse msg as []interface{} if needed
			inputs := []interface{}{string(msg)} // Currently just a string

			_ = action.SetInputs(inputs)

			result := action.Run()(action)

			x_log.Info("Action executed. Result:", result)
		}

		err = bm.busClient.SubscribeToTopic(subjectWithMethod, "", handler)
		if err != nil {
			x_log.Error("Failed to subscribe to subject", subjectWithMethod, ":", err)
		} else {
			x_log.Info("Subscribed to subject:", subjectWithMethod)
		}
	}
}
