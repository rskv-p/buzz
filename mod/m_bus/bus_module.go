// file: buzz/mod/m_bus/bus_module.go

package m_bus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	mod.Module               // <-- теперь указатель на Module
	Bus        *bus_core.Bus // Embedded bus (message processing)
	busClient  typ.IBusClient
}

// NewBusModule creates and returns a new BusModule with all actions registered.
func NewBusModule(name string, secretKey string, selfClient typ.IBusClient, maxGoroutines int) *BusModule {
	bus := bus_core.NewBus(secretKey, selfClient, maxGoroutines)

	busClient := bus_client.NewBusClient(1, bus, secretKey, 10) // ID = 1, можно подставить другое

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

// ProcessRequestBody decodes the JSON body of a request into a list of inputs.
func ProcessRequestBody(r *http.Request) ([]interface{}, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("request body is missing")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Log the body for debugging purposes
	x_log.Debug("Request body:", string(body))

	// Recreate the request body since it has been consumed
	r.Body = io.NopCloser(bytes.NewReader(body))

	var inputs []interface{}
	if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}
	return inputs, nil
}

// SetActionInputs safely sets the inputs for the action and handles any potential errors.
func SetActionInputs(action typ.IAction, inputs []interface{}) error {
	if action == nil {
		return fmt.Errorf("action is nil")
	}

	if err := action.SetInputs(inputs); err != nil {
		return fmt.Errorf("failed to set inputs: %w", err)
	}
	return nil
}

// HandleAction performs the action and returns the result.
func HandleAction(action typ.IAction) (any, error) {
	if action == nil {
		return nil, fmt.Errorf("action is nil")
	}

	return action.Run()(action), nil
}

// RegisterPublicActions registers actions that are public and available via REST API.
func RegisterPublicActions(module typ.IModule) {
	actions := module.GetActions()

	for _, action := range actions {
		if !action.IsPublic() {
			continue
		}

		actionName := action.GetName()
		expectedMethods := action.GetMethod()

		if len(expectedMethods) == 0 {
			x_log.Warn("No HTTP methods specified for action:", actionName)
			continue
		}

		// Копируем action в локальную переменную, чтобы избежать замыкания
		act := action

		// Регистрируем один обработчик на путь
		http.HandleFunc("/api/"+actionName, func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что метод разрешён
			if r.Method != expectedMethods {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			// Логируем запрос
			x_log.Info("Received request:", r.Method, "for action:", act.GetName())

			switch r.Method {
			case http.MethodGet:
				handleGetRequest(w, r, act)
			case http.MethodPost:
				handlePostRequest(w, r, act)
			case http.MethodPut:
				handlePutRequest(w, r, act)
			case http.MethodDelete:
				handleDeleteRequest(w, r, act)
			default:
				http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
			}
		})

		x_log.Info("Published REST API for action:", actionName)
	}
}

// handleGetRequest processes GET requests.
func handleGetRequest(w http.ResponseWriter, _ *http.Request, action typ.IAction) {
	// Execute the action and send the result
	result, err := HandleAction(action)
	if err != nil {
		x_log.Error("Error handling action:", err)
		http.Error(w, "Action execution failed", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	sendJSONResponse(w, result)
}

// handlePostRequest processes POST requests.
func handlePostRequest(w http.ResponseWriter, r *http.Request, action typ.IAction) {
	// Process the POST request body
	inputs, err := ProcessRequestBody(r)
	if err != nil {
		x_log.Error("Error decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set the inputs for the action
	if err := SetActionInputs(action, inputs); err != nil {
		x_log.Error("Error setting inputs:", err)
		http.Error(w, "Failed to set action inputs", http.StatusInternalServerError)
		return
	}

	// Execute the action and send the result
	result, err := HandleAction(action)
	if err != nil {
		x_log.Error("Error handling action:", err)
		http.Error(w, "Action execution failed", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	sendJSONResponse(w, result)
}

// handlePutRequest processes PUT requests.
func handlePutRequest(w http.ResponseWriter, r *http.Request, action typ.IAction) {
	// Similar to POST but with different semantics for PUT
	inputs, err := ProcessRequestBody(r)
	if err != nil {
		x_log.Error("Error decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := SetActionInputs(action, inputs); err != nil {
		x_log.Error("Error setting inputs:", err)
		http.Error(w, "Failed to set action inputs", http.StatusInternalServerError)
		return
	}

	result, err := HandleAction(action)
	if err != nil {
		x_log.Error("Error handling action:", err)
		http.Error(w, "Action execution failed", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, result)
}

// handleDeleteRequest processes DELETE requests.
func handleDeleteRequest(w http.ResponseWriter, _ *http.Request, action typ.IAction) {
	// Perform the DELETE action
	result, err := HandleAction(action)
	if err != nil {
		x_log.Error("Error handling action:", err)
		http.Error(w, "Action execution failed", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, result)
}

// sendJSONResponse is a helper function to send JSON responses to the client.
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		x_log.Error("Failed to send JSON response:", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

func (bm *BusModule) RegisterActionsOnBus() {
	for _, action := range bm.Module.GetActions() {
		if !action.IsPublic() {
			continue
		}

		// Получаем базовое имя action
		baseSubject := action.GetName()
		// Добавляем метод для шины
		subjectWithMethod, err := x_str.BuildTopicName(baseSubject, action.GetMethod())
		if err != nil {
			x_log.Error("Failed to build full bus subject for action:", baseSubject, "Error:", err)
			continue
		}

		handler := func(subject string, msg []byte) {
			x_log.Info("Executing action for subject:", subject)

			// Здесь можно парсить msg как []interface{}, если надо
			inputs := []interface{}{string(msg)} // Пока просто строка

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
