// file: buzz/mod/m_bus/bus_req/request_test.go

package bus_req_test

import (
	"testing"

	"github.com/rskv-p/buzz/mod/m_bus/bus_req"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHandler is used to mock Respond and Error handling functions
type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) Respond(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockHandler) RespJSON(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockHandler) Err(code, msg string, data interface{}) error {
	args := m.Called(code, msg, data)
	return args.Error(0)
}

// TestRequest tests the Request struct and its methods
func TestRequest(t *testing.T) {
	// Create a mock handler for Respond, RespJSON, and Err
	mockHandler := new(MockHandler)

	// Create a new Request for testing
	req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

	// Test RespondJSON method
	t.Run("Test RespondJSON", func(t *testing.T) {
		// Set up mock behavior for the Respond method
		mockHandler.On("Respond", mock.Anything).Return(nil)
		req.Respond = mockHandler.Respond

		// Create a test response object
		response := map[string]string{"status": "success"}

		// Call RespondJSON
		err := req.RespondJSON(response)
		assert.NoError(t, err, "Expected no error while responding with JSON")

		// Check that the mock Respond method was called
		mockHandler.AssertExpectations(t)
	})

	// Test Error method with custom error handler
	t.Run("Test Error with Custom Handler", func(t *testing.T) {
		// Set up mock behavior for the custom error handler
		mockHandler.On("Err", "404", "Not Found", mock.Anything).Return(nil)
		req.SetErrorHandler(mockHandler.Err)

		// Call Error with expected "404" code and message
		err := req.Error("404", "Not Found", nil)
		assert.NoError(t, err, "Expected no error while sending error")

		// Check that the mock Err method was called
		mockHandler.AssertExpectations(t)
	})

	// Test Error method with default handler
	t.Run("Test Error with Default Handler", func(t *testing.T) {
		// Set up mock behavior for the custom error handler (use the same values as above)
		mockHandler.On("Err", "500", "Internal Server Error", mock.Anything).Return(nil)
		req.SetErrorHandler(mockHandler.Err)

		// Call Error with expected "500" code and message
		err := req.Error("500", "Internal Server Error", nil)
		assert.NoError(t, err, "Expected no error while sending error")

		// Ensure the mock Err method was called
		mockHandler.AssertExpectations(t)
	})

	// Test SetHeader and Headers methods
	t.Run("Test SetHeader and Headers", func(t *testing.T) {
		// Set a header
		err := req.SetHeader("Content-Type", "application/json")
		assert.NoError(t, err, "Expected no error while setting header")

		// Get the headers
		headers, err := req.Headers()
		assert.NoError(t, err, "Expected no error while getting headers")

		// Check that the header is correctly set
		value, _ := headers.Get("Content-Type")
		assert.Equal(t, "application/json", value, "Expected header value to be 'application/json'")
	})

	// Test GetSubject and GetData methods
	t.Run("Test GetSubject and GetData", func(t *testing.T) {
		// Test subject
		subject := req.GetSubject()
		assert.Equal(t, "test.topic", subject, "Expected subject to be 'test.topic'")

		// Test data
		data := req.GetData()
		assert.Equal(t, []byte("Test data"), data, "Expected data to be 'Test data'")
	})
}
