package m_gate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rskv-p/buzz/act"
	"github.com/rskv-p/buzz/mod"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

var _ typ.IModule = (*GateModule)(nil) // Ensure RequestHeaders implements IRequestHeaders

//---------------------
// Error Handling
//---------------------

var ErrInvalidActionType = fmt.Errorf("invalid action type")

// GateModule combines Module and BusClient together
type GateModule struct {
	mod.Module // Embedding the base Module
	busClient  typ.IBusClient
}

// NewGateModule creates a new GateModule with actions.
func NewGateModule(name string, busClient typ.IBusClient) *GateModule {
	// Define actions (example, add more actions as needed)
	actions := map[string]typ.IAction{
		"action_name": &act.Action{
			Name:   "action_name",
			Public: true,
			Func:   func(typ.IAction) any { return "result" },
			Method: http.MethodGet,
		},
	}

	// Create the module
	module := &mod.Module{
		Name:    name,
		Actions: actions,
	}

	return &GateModule{
		Module:    *module,
		busClient: busClient,
	}
}

// RegisterPublicActions registers actions that are public.
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

		// Register one handler for the path
		http.HandleFunc("/api/"+actionName, createActionHandler(action, expectedMethods))

		x_log.Info("Published REST API for action:", actionName)
	}
}

// createActionHandler creates an HTTP handler for the action
func createActionHandler(action typ.IAction, expectedMethods string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the method is allowed
		if r.Method != expectedMethods {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Log the request
		x_log.Info("Received request:", r.Method, "for action:", action.GetName())

		var result any
		var err error

		// Process the request based on the method
		switch r.Method {
		case http.MethodGet:
			result, err = handleGetRequest(action)
		case http.MethodPost:
			result, err = handlePostRequest(w, r, action)
		case http.MethodPut:
			result, err = handlePutRequest(w, r, action)
		case http.MethodDelete:
			result, err = handleDeleteRequest(action)
		default:
			http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
			return
		}

		// Handle any errors and send response
		if err != nil {
			x_log.Error("Error handling action:", err)
			http.Error(w, "Action execution failed", http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, result)
	}
}

// handleGetRequest processes GET requests.
func handleGetRequest(action typ.IAction) (any, error) {
	// Execute the action and send the result
	return HandleAction(action)
}

// handlePostRequest processes POST requests.
func handlePostRequest(_ http.ResponseWriter, r *http.Request, action typ.IAction) (any, error) {
	// Process the POST request body
	inputs, err := ProcessRequestBody(r)
	if err != nil {
		return nil, fmt.Errorf("error decoding request body: %w", err)
	}

	// Set the inputs for the action
	if err := SetActionInputs(action, inputs); err != nil {
		return nil, fmt.Errorf("error setting inputs: %w", err)
	}

	// Execute the action
	return HandleAction(action)
}

// handlePutRequest processes PUT requests.
func handlePutRequest(_ http.ResponseWriter, r *http.Request, action typ.IAction) (any, error) {
	// Similar to POST but with different semantics for PUT
	inputs, err := ProcessRequestBody(r)
	if err != nil {
		return nil, fmt.Errorf("error decoding request body: %w", err)
	}

	if err := SetActionInputs(action, inputs); err != nil {
		return nil, fmt.Errorf("error setting inputs: %w", err)
	}

	// Execute the action
	return HandleAction(action)
}

// handleDeleteRequest processes DELETE requests.
func handleDeleteRequest(action typ.IAction) (any, error) {
	// Perform the DELETE action
	return HandleAction(action)
}

// sendJSONResponse sends a JSON response to the client.
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		x_log.Error("Failed to send JSON response:", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
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
