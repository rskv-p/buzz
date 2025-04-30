// file: buzz/mod/m_bus/bus_req/request_test.go

package bus_req_test

import (
	"testing"

	"github.com/rskv-p/buzz/mod/m_bus/bus_req"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//-----------------------------------------
//  MockHandler
//-----------------------------------------

// MockHandler mocks response and error handlers for Request.
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

//-----------------------------------------
//  Request Tests
//-----------------------------------------

func TestRequest(t *testing.T) {
	x_init.Init()

	mockHandler := new(MockHandler)
	req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

	//-----------------------------------------
	//  RespondJSON
	//-----------------------------------------
	t.Run("Test RespondJSON", func(t *testing.T) {
		mockHandler.On("Respond", mock.Anything).Return(nil)
		req.Respond = mockHandler.Respond

		response := map[string]string{"status": "success"}
		err := req.RespondJSON(response)

		assert.NoError(t, err)
		mockHandler.AssertExpectations(t)
	})

	//-----------------------------------------
	//  Error with Custom Handler
	//-----------------------------------------
	t.Run("Test Error with Custom Handler", func(t *testing.T) {
		mockHandler.On("Err", "404", "Not Found", mock.Anything).Return(nil)
		req.SetErrorHandler(mockHandler.Err)

		err := req.Error("404", "Not Found", nil)

		assert.NoError(t, err)
		mockHandler.AssertExpectations(t)
	})

	//-----------------------------------------
	//  Error with Default Handler
	//-----------------------------------------
	t.Run("Test Error with Default Handler", func(t *testing.T) {
		mockHandler.On("Err", "500", "Internal Server Error", mock.Anything).Return(nil)
		req.SetErrorHandler(mockHandler.Err)

		err := req.Error("500", "Internal Server Error", nil)

		assert.NoError(t, err)
		mockHandler.AssertExpectations(t)
	})

	//-----------------------------------------
	//  SetHeader and Headers
	//-----------------------------------------
	t.Run("Test SetHeader and Headers", func(t *testing.T) {
		err := req.SetHeader("Content-Type", "application/json")
		assert.NoError(t, err)

		headers, err := req.Headers()
		assert.NoError(t, err)

		value, _ := headers.Get("Content-Type")
		assert.Equal(t, "application/json", value)
	})

	//-----------------------------------------
	//  GetSubject and GetData
	//-----------------------------------------
	t.Run("Test GetSubject and GetData", func(t *testing.T) {
		assert.Equal(t, "test.topic", req.GetSubject())
		assert.Equal(t, []byte("Test data"), req.GetData())
	})
}
