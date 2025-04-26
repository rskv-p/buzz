// file: buz/bus/bus_req/request.go

package bus_req

import (
	"encoding/json"
	"fmt"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_type"
)

//-------------------------------------------------
// Request - Represents an incoming request with associated handlers and metadata
//-------------------------------------------------

// Request represents an incoming request with associated handlers and metadata.
type Request struct {
	Subject string
	Reply   string
	Data    []byte
	headers bus_type.IRequestHeaders

	Respond  func([]byte) error
	RespJSON func(interface{}) error
	Err      func(code, msg string, data interface{}) error
}

//-------------------------------------------------
// RespondJSON - Sends a JSON-encoded response
//-------------------------------------------------

// RespondJSON sends a JSON-encoded response to the Reply address.
func (r *Request) RespondJSON(v interface{}) error {
	bus_comm.Infof("Responding to subject %s with JSON", r.Subject)

	// Marshal the value into JSON
	data, err := json.Marshal(v)
	if err != nil {
		bus_comm.Errorf("Failed to marshal JSON for subject %s: %v", r.Subject, err)
		return fmt.Errorf("marshal error: %w", err)
	}

	// Send the response
	if err := r.Respond(data); err != nil {
		bus_comm.Errorf("Failed to send response for subject %s: %v", r.Subject, err)
		return err
	}

	// Log success
	bus_comm.Infof("Response sent successfully for subject %s", r.Subject)
	return nil
}

//-------------------------------------------------
// Error - Sends a standard error payload as JSON
//-------------------------------------------------

// Error sends a standard error payload as JSON.
func (r *Request) Error(code, description string, data []byte) error {
	bus_comm.Infof("Sending error for subject %s with code %s and description %s", r.Subject, code, description)

	if r.Err != nil {
		bus_comm.Errorf("Custom error handler for subject %s with code %s and description %s", r.Subject, code, description)
		return r.Err(code, description, data)
	}

	// Send a default error response
	errorResponse := map[string]string{
		"error":       code,
		"description": description,
	}
	bus_comm.Errorf("Sending default error response for subject %s with code %s and description %s", r.Subject, code, description)
	return r.RespondJSON(errorResponse)
}

//-------------------------------------------------
// NewTestRequest - Creates a mock Request for testing
//-------------------------------------------------

// NewTestRequest creates a mock Request for testing purposes.
func NewTestRequest(subject string, data []byte) *Request {
	bus_comm.Infof("Creating a new test request for subject %s", subject)
	return &Request{
		Subject:  subject,
		Data:     data,
		headers:  &RequestHeaders{Headers: make(map[string][]string)}, // Initialize headers with an empty map
		Respond:  func(_ []byte) error { return nil },
		RespJSON: func(_ interface{}) error { return nil },
		Err: func(code, msg string, _ interface{}) error {
			bus_comm.Infof("Mock error handler for subject %s", subject)
			return nil
		},
	}
}

//-------------------------------------------------
// SetErrorHandler - Sets a custom error handler
//-------------------------------------------------

// SetErrorHandler sets a custom error handler for the request.
func (r *Request) SetErrorHandler(f func(code, msg string, data interface{}) error) {
	bus_comm.Infof("Setting custom error handler for subject %s", r.Subject)
	r.Err = f
}

//-------------------------------------------------
// SetHeader - Sets a header for the request
//-------------------------------------------------

// SetHeader sets a header for the request.
func (r *Request) SetHeader(key, value string) error {
	if key == "" || value == "" {
		bus_comm.Errorf("Key or value cannot be empty for subject %s", r.Subject)
		return fmt.Errorf("key or value cannot be empty")
	}

	if r.headers == nil {
		bus_comm.Infof("Initializing headers for subject %s", r.Subject)
		r.headers = &RequestHeaders{Headers: make(map[string][]string)}
	}

	r.headers.Set(key, value)
	bus_comm.Infof("Header set for subject %s with key %s and value %s", r.Subject, key, value)
	return nil
}

//-------------------------------------------------
// Headers - Returns the headers associated with the request
//-------------------------------------------------

// Headers returns the headers associated with the request.
func (r *Request) Headers() (bus_type.IRequestHeaders, error) {
	if r.headers == nil {
		bus_comm.Errorf("Headers are not initialized for subject %s", r.Subject)
		return nil, fmt.Errorf("headers are not initialized")
	}
	bus_comm.Infof("Returning headers for subject %s", r.Subject)
	return r.headers, nil
}

//-------------------------------------------------
// GetSubject - Returns the subject of the request
//-------------------------------------------------

// GetSubject retrieves the subject of the request.
func (r *Request) GetSubject() string {
	bus_comm.Infof("Getting subject for request: %s", r.Subject)
	return r.Subject
}

//-------------------------------------------------
// GetData - Returns the data of the request
//-------------------------------------------------

// GetData retrieves the data of the request.
func (r *Request) GetData() []byte {
	bus_comm.Infof("Getting data for request with subject %s", r.Subject)
	return r.Data
}

//-------------------------------------------------
// GetReply - Returns the reply address of the request
//-------------------------------------------------

// GetReply retrieves the reply address for the request.
func (r *Request) GetReply() string {
	bus_comm.Infof("Getting reply for request with subject %s", r.Subject)
	return r.Reply
}
