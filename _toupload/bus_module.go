// file: buzz/mod/m_bus/bus_module.go

package bus

import (
	"fmt"

	"github.com/rskv-p/buzz/act"
	"github.com/rskv-p/buzz/mod"
	"github.com/rskv-p/buzz/mod/m_bus/bus_core"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

//---------------------
// Error Handling
//---------------------

// ErrInvalidActionType is the error returned when an action cannot be cast to *act.Action.
var ErrInvalidActionType = fmt.Errorf("invalid action type")

//---------------------
// Bus Module
//---------------------

// BusModule creates and returns a bus module with actions for publish, subscribe, and stats.
func BusModule(bus *bus_core.Bus) typ.IModule {
	return &mod.Module{
		Name: "bus",
		Actions: []typ.ActionDef{
			{
				Name:   "bus.demo",
				Func:   HandleDemo,
				Public: true,
			},
		},
		OnInit: func() error {
			// Initialize the bus module (no specific initialization required here)
			return nil
		},
		OnStop: func() error {
			// Stop the bus module (no specific stopping required here)
			return nil
		},
	}
}

//---------------------
// Helper Functions
//---------------------

// handleActionCast safely casts IAction to *act.Action and handles errors.
func handleActionCast(a typ.IAction) (*act.Action, error) {
	action, ok := a.(*act.Action)
	if !ok {
		return nil, ErrInvalidActionType
	}
	return action, nil
}

//---------------------
// Handlers for Actions
//---------------------

// HandlePublish processes the action to publish a message through the bus.
func HandleDemo(a typ.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		x_log.Error("Failed to cast action to *act.Action:", err)
		return err
	}

	subject := action.InputString(0)
	msg, ok := action.Inputs[1].([]byte)
	if !ok {
		return fmt.Errorf("invalid message type for publish: expected []byte")
	}

	x_log.Info("Successfully published message to subject: ", subject, "msg: ", msg)
	return true
}
