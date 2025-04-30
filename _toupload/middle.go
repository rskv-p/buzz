// file: buzz/mod/m_bus/bus_middle/middle.go

package bus_middle

import (
	"fmt"

	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

//-------------------------------------------------
// Middleware - Represents a base middleware implementation
//-------------------------------------------------

// Middleware represents a base middleware implementation.
type Middleware struct {
	Handler func(typ.IRequest) error // The handler function for processing requests
}

//-------------------------------------------------
// Process - Executes the middleware logic and logs the details
//-------------------------------------------------

// Process executes the middleware logic and logs the details.
func (m *Middleware) Process(req typ.IRequest) error {
	// Log incoming request details
	x_log.Debug("Processing request", req)

	// Check if handler is defined
	if m.Handler != nil {
		// Try to process the request and log the outcome
		err := m.Handler(req)
		if err != nil {
			// Log the error in case of failure
			x_log.Error("Error processing request", req, "Error", err)
			return fmt.Errorf("middleware failed to process request: %v, error: %v", req, err)
		}

		// Log successful processing of the request
		x_log.Info("Successfully processed request", req)
	} else {
		// Log if no handler is defined
		x_log.Warn("No handler defined for request")
	}

	return nil
}

//-------------------------------------------------
// NewMiddleware - Creates a new middleware instance
//-------------------------------------------------

// NewMiddleware creates a new middleware instance.
func NewMiddleware(handler func(typ.IRequest) error) typ.IMiddleware {
	return &Middleware{Handler: handler}
}
