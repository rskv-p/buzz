// file: buz/bus/bus_middle/middle_test.go

package bus_middle_test

import (
	"fmt"
	"testing"

	"github.com/rskv-p/buzz/bus/bus_middle"
	"github.com/rskv-p/buzz/bus/bus_req"
	"github.com/rskv-p/buzz/bus/bus_type"
	"github.com/stretchr/testify/assert"
)

// TestMiddleware tests the middleware logic
func TestMiddleware(t *testing.T) {
	// Test with a handler that succeeds
	t.Run("Test Process with Successful Handler", func(t *testing.T) {
		// Create a real Request object
		req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

		// Create a middleware with a handler that does not return an error
		middleware := bus_middle.NewMiddleware(func(req bus_type.IRequest) error {
			// Validate subject and data in the request
			assert.Equal(t, "test.topic", req.GetSubject(), "Expected subject to be 'test.topic'")
			assert.Equal(t, []byte("Test data"), req.GetData(), "Expected data to be 'Test data'")
			return nil
		})

		// Process the request
		err := middleware.Process(req)
		assert.NoError(t, err, "Expected no error during processing")
	})

	// Test with a handler that returns an error
	t.Run("Test Process with Handler Returning Error", func(t *testing.T) {
		// Create a real Request object
		req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

		// Create a middleware with a handler that returns an error
		middleware := bus_middle.NewMiddleware(func(req bus_type.IRequest) error {
			return fmt.Errorf("handler error")
		})

		// Process the request
		err := middleware.Process(req)
		assert.Error(t, err, "Expected an error during processing")
	})

	// Test with no handler
	t.Run("Test Process with No Handler", func(t *testing.T) {
		// Create a real Request object
		req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

		// Create a middleware with no handler (nil)
		middleware := bus_middle.NewMiddleware(nil)

		// Process the request
		err := middleware.Process(req)
		assert.NoError(t, err, "Expected no error during processing when handler is not defined")
	})
}
