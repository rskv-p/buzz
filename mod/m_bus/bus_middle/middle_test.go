// file: buzz/mod/m_bus/bus_middle/middle_test.go

package bus_middle_test

import (
	"fmt"
	"testing"

	"github.com/rskv-p/buzz/mod/m_bus/bus_middle"
	"github.com/rskv-p/buzz/mod/m_bus/bus_req"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/rskv-p/buzz/typ"
	"github.com/stretchr/testify/assert"
)

//-----------------------------------------
//  Middleware Tests
//-----------------------------------------

func TestMiddleware(t *testing.T) {
	x_init.Init()

	//-----------------------------------------
	//  Process with Successful Handler
	//-----------------------------------------
	t.Run("Test Process with Successful Handler", func(t *testing.T) {
		req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

		middleware := bus_middle.NewMiddleware(func(req typ.IRequest) error {
			assert.Equal(t, "test.topic", req.GetSubject())
			assert.Equal(t, []byte("Test data"), req.GetData())
			return nil
		})

		err := middleware.Process(req)
		assert.NoError(t, err)
	})

	//-----------------------------------------
	//  Process with Handler Returning Error
	//-----------------------------------------
	t.Run("Test Process with Handler Returning Error", func(t *testing.T) {
		req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

		middleware := bus_middle.NewMiddleware(func(req typ.IRequest) error {
			return fmt.Errorf("handler error")
		})

		err := middleware.Process(req)
		assert.Error(t, err)
	})

	//-----------------------------------------
	//  Process with No Handler
	//-----------------------------------------
	t.Run("Test Process with No Handler", func(t *testing.T) {
		req := bus_req.NewTestRequest("test.topic", []byte("Test data"))

		middleware := bus_middle.NewMiddleware(nil)

		err := middleware.Process(req)
		assert.NoError(t, err)
	})
}
