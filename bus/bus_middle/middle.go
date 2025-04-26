// file: buz/bus/bus_middle/middle.go

package bus_middle

import (
	"fmt"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_type"
)

//-------------------------------------------------
// Middleware - Represents a base middleware implementation
//-------------------------------------------------

// Middleware represents a base middleware implementation.
type Middleware struct {
	Handler func(bus_type.IRequest) error // The handler function for processing requests
}

//-------------------------------------------------
// Process - Executes the middleware logic and logs the details
//-------------------------------------------------

// Process executes the middleware logic and logs the details.
func (m *Middleware) Process(req bus_type.IRequest) error {
	// Log incoming request details
	bus_comm.Debugf("Processing request: %v", req)

	// Check if handler is defined
	if m.Handler != nil {
		// Try to process the request and log the outcome
		err := m.Handler(req)
		if err != nil {
			// Log the error in case of failure
			bus_comm.Errorf("Error processing request: %v, Error: %v", req, err)
			return fmt.Errorf("middleware failed to process request: %v, error: %v", req, err)
		}

		// Log successful processing of the request
		bus_comm.Infof("Successfully processed request: %v", req)
	} else {
		// Log if no handler is defined
		bus_comm.Warn("No handler defined for request")
	}

	return nil
}

//-------------------------------------------------
// NewMiddleware - Creates a new middleware instance
//-------------------------------------------------

// NewMiddleware creates a new middleware instance.
func NewMiddleware(handler func(bus_type.IRequest) error) bus_type.IMiddleware {
	return &Middleware{Handler: handler}
}
