// file: buzz/mod/m_bus/bus_req/headers_test.go

package bus_req_test

import (
	"testing"

	"github.com/rskv-p/buzz/mod/m_bus/bus_req"
	"github.com/stretchr/testify/assert"
)

// TestRequestHeaders tests the methods of the RequestHeaders struct
func TestRequestHeaders(t *testing.T) {
	// Create a new RequestHeaders instance
	headers := &bus_req.RequestHeaders{Headers: make(map[string][]string)}

	// Test Set method
	t.Run("Test Set", func(t *testing.T) {
		err := headers.Set("Content-Type", "application/json")
		assert.NoError(t, err, "Expected no error while setting header")

		// Verify the header is set correctly
		value, err := headers.Get("Content-Type")
		assert.NoError(t, err, "Expected no error while getting header")
		assert.Equal(t, "application/json", value, "Expected header value to be 'application/json'")
	})

	// Test Get method
	t.Run("Test Get", func(t *testing.T) {
		value, err := headers.Get("Content-Type")
		assert.NoError(t, err, "Expected no error while getting header")
		assert.Equal(t, "application/json", value, "Expected header value to be 'application/json'")

		// Test for a non-existent header
		_, err = headers.Get("Non-Existent")
		assert.Error(t, err, "Expected error for non-existent header")
	})

	// Test Remove method
	t.Run("Test Remove", func(t *testing.T) {
		// Set header first
		err := headers.Set("Authorization", "Bearer token")
		assert.NoError(t, err, "Expected no error while setting header")

		// Remove header
		err = headers.Remove("Authorization")
		assert.NoError(t, err, "Expected no error while removing header")

		// Try to get the removed header
		_, err = headers.Get("Authorization")
		assert.Error(t, err, "Expected error for removed header")
	})

	// Test GetAll method
	t.Run("Test GetAll", func(t *testing.T) {
		// Set multiple headers
		err := headers.Set("X-Custom-Header", "custom-value")
		assert.NoError(t, err, "Expected no error while setting header")

		// Get all headers
		allHeaders := headers.GetAll()
		assert.Equal(t, 2, len(allHeaders), "Expected 2 headers")
		assert.Equal(t, "application/json", allHeaders["Content-Type"][0], "Expected Content-Type to be 'application/json'")
		assert.Equal(t, "custom-value", allHeaders["X-Custom-Header"][0], "Expected X-Custom-Header to be 'custom-value'")
	})
}
