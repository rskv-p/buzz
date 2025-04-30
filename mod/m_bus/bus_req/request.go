// file: buzz/mod/m_bus/bus_req/request.go

package bus_req

import (
	"encoding/json"
	"fmt"

	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

//-------------------------------------------------
// Request - Represents an incoming request with associated handlers and metadata
//-------------------------------------------------

// Request represents an incoming request with associated handlers and metadata.
type Request struct {
	Subject string
	Reply   string
	Data    []byte
	headers typ.IRequestHeaders

	Respond  func([]byte) error
	RespJSON func(interface{}) error
	Err      func(code, msg string, data interface{}) error
}

//-------------------------------------------------
// RespondJSON - Sends a JSON-encoded response
//-------------------------------------------------

// RespondJSON sends a JSON-encoded response to the Reply address.
func (r *Request) RespondJSON(v interface{}) error {
	x_log.Info("Responding to subject", r.Subject, "with JSON")

	// Marshal the value into JSON
	data, err := json.Marshal(v)
	if err != nil {
		x_log.Error("Failed to marshal JSON for subject", r.Subject, "error", err)
		return fmt.Errorf("marshal error: %w", err)
	}

	// Send the response
	if err := r.Respond(data); err != nil {
		x_log.Error("Failed to send response for subject", r.Subject, "error", err)
		return err
	}

	// Log success
	x_log.Info("Response sent successfully for subject", r.Subject)
	return nil
}

//-------------------------------------------------
// Error - Sends a standard error payload as JSON
//-------------------------------------------------

// Error sends a standard error payload as JSON.
func (r *Request) Error(code, description string, data []byte) error {
	x_log.Info("Sending error for subject", r.Subject, "code", code, "description", description)

	if r.Err != nil {
		x_log.Error("Custom error handler for subject", r.Subject, "code", code, "description", description)
		return r.Err(code, description, data)
	}

	// Send a default error response
	errorResponse := map[string]string{
		"error":       code,
		"description": description,
	}
	x_log.Error("Sending default error response for subject", r.Subject, "code", code, "description", description)
	return r.RespondJSON(errorResponse)
}

//-------------------------------------------------
// NewTestRequest - Creates a mock Request for testing
//-------------------------------------------------

// NewTestRequest creates a mock Request for testing purposes.
func NewTestRequest(subject string, data []byte) *Request {
	x_log.Info("Creating a new test request for subject", subject)
	return &Request{
		Subject:  subject,
		Data:     data,
		headers:  &RequestHeaders{Headers: make(map[string][]string)}, // Initialize headers with an empty map
		Respond:  func(_ []byte) error { return nil },
		RespJSON: func(_ interface{}) error { return nil },
		Err: func(code, msg string, _ interface{}) error {
			x_log.Info("Mock error handler for subject", subject)
			return nil
		},
	}
}

//-------------------------------------------------
// SetErrorHandler - Sets a custom error handler
//-------------------------------------------------

// SetErrorHandler sets a custom error handler for the request.
func (r *Request) SetErrorHandler(f func(code, msg string, data interface{}) error) {
	x_log.Info("Setting custom error handler for subject", r.Subject)
	r.Err = f
}

//-------------------------------------------------
// SetHeader - Sets a header for the request
//-------------------------------------------------

// SetHeader sets a header for the request.
func (r *Request) SetHeader(key, value string) error {
	if key == "" || value == "" {
		x_log.Error("Key or value cannot be empty for subject", r.Subject)
		return fmt.Errorf("key or value cannot be empty")
	}

	if r.headers == nil {
		x_log.Info("Initializing headers for subject", r.Subject)
		r.headers = &RequestHeaders{Headers: make(map[string][]string)}
	}

	r.headers.Set(key, value)
	x_log.Info("Header set for subject", r.Subject, "key", key, "value", value)
	return nil
}

//-------------------------------------------------
// Headers - Returns the headers associated with the request
//-------------------------------------------------

// Headers returns the headers associated with the request.
func (r *Request) Headers() (typ.IRequestHeaders, error) {
	if r.headers == nil {
		x_log.Error("Headers are not initialized for subject", r.Subject)
		return nil, fmt.Errorf("headers are not initialized")
	}
	x_log.Info("Returning headers for subject", r.Subject)
	return r.headers, nil
}

//-------------------------------------------------
// GetSubject - Returns the subject of the request
//-------------------------------------------------

// GetSubject retrieves the subject of the request.
func (r *Request) GetSubject() string {
	x_log.Info("Getting subject for request", r.Subject)
	return r.Subject
}

//-------------------------------------------------
// GetData - Returns the data of the request
//-------------------------------------------------

// GetData retrieves the data of the request.
func (r *Request) GetData() []byte {
	x_log.Info("Getting data for request with subject", r.Subject)
	return r.Data
}

//-------------------------------------------------
// GetReply - Returns the reply address of the request
//-------------------------------------------------

// GetReply retrieves the reply address for the request.
func (r *Request) GetReply() string {
	x_log.Info("Getting reply for request with subject", r.Subject)
	return r.Reply
}
