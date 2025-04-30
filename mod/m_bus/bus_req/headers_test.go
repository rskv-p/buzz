package bus_req_test

import (
	"testing"

	"github.com/rskv-p/buzz/mod/m_bus/bus_req"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/stretchr/testify/assert"
)

//-----------------------------------------
//  RequestHeaders Tests
//-----------------------------------------

func TestRequestHeaders(t *testing.T) {
	x_init.Init()

	headers := &bus_req.RequestHeaders{Headers: make(map[string][]string)}

	//-----------------------------------------
	//  Set Header
	//-----------------------------------------
	t.Run("Test Set", func(t *testing.T) {
		err := headers.Set("Content-Type", "application/json")
		assert.NoError(t, err)

		value, err := headers.Get("Content-Type")
		assert.NoError(t, err)
		assert.Equal(t, "application/json", value)
	})

	//-----------------------------------------
	//  Get Header
	//-----------------------------------------
	t.Run("Test Get", func(t *testing.T) {
		value, err := headers.Get("Content-Type")
		assert.NoError(t, err)
		assert.Equal(t, "application/json", value)

		_, err = headers.Get("Non-Existent")
		assert.Error(t, err)
	})

	//-----------------------------------------
	//  Remove Header
	//-----------------------------------------
	t.Run("Test Remove", func(t *testing.T) {
		err := headers.Set("Authorization", "Bearer token")
		assert.NoError(t, err)

		err = headers.Remove("Authorization")
		assert.NoError(t, err)

		_, err = headers.Get("Authorization")
		assert.Error(t, err)
	})

	//-----------------------------------------
	//  Get All Headers
	//-----------------------------------------
	t.Run("Test GetAll", func(t *testing.T) {
		err := headers.Set("X-Custom-Header", "custom-value")
		assert.NoError(t, err)

		allHeaders := headers.GetAll()
		assert.Equal(t, 2, len(allHeaders))
		assert.Equal(t, "application/json", allHeaders["Content-Type"][0])
		assert.Equal(t, "custom-value", allHeaders["X-Custom-Header"][0])
	})
}
