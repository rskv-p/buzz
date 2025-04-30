// file: buzz/act/action.go

package act

import (
	"context"
	"fmt"
	"strings"

	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

var _ typ.IAction = (*Action)(nil)

//---------------------
// Types
//---------------------

type Action struct {
	ID      string // Resource ID
	Name    string // Full name (e.g. "models.user.get")
	Method  string // Method name
	Handler string // Registered handler key
	Inputs  []any  // Input arguments
	output  *any   // Execution result
	ctx     context.Context
	Bus     typ.IBusClient // Access to the bus
	module  typ.IModule
	Public  bool // Public now a bool
	Func    func(typ.IAction) any
}

//---------------------
// Factory
//---------------------

// NewAction creates a new Action, optionally with a context.
func NewAction(name string, ctx context.Context, args ...any) (*Action, error) {
	action, err := createAction(name, args...)
	if err != nil {
		x_log.Info("Error creating action:", err)
		return nil, fmt.Errorf("action: %w", err)
	}

	// Set context if provided
	if ctx != nil {
		action = action.WithContext(ctx).(*Action)
	}
	return action, nil
}

// createAction is a helper function to create an Action without the context part.
func createAction(name string, args ...any) (*Action, error) {
	act := &Action{Name: name, Inputs: args}
	return act.init() // returns initialized Action
}

//---------------------
// Builders
//---------------------

// GetFunc returns the Func handler function for the Action, returning a function of type func(typ.IAction) any.
func (a *Action) Run() func(typ.IAction) any {
	return a.Func
}

func (a *Action) WithContext(ctx context.Context) typ.IAction {
	a.ctx = ctx
	return a
}

// GetName returns the name of the Action.
func (a *Action) GetName() string {
	return a.Name
}

// GetName returns the name of the Action.
func (a *Action) GetModuleName() string {
	return a.module.GetName()
}

// GetName returns the name of the Action.
func (a *Action) GetMethod() string {
	return a.Method
}

// IsPublic returns the value of the Public field (now a bool).
func (a *Action) IsPublic() bool {
	return a.Public
}

func (a *Action) Context() context.Context {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	return a.ctx
}

func (a *Action) Release() {
	a.output = nil
}

func (a *Action) Dispose() {
	a.Inputs = nil
	a.output = nil
	a.ctx = nil
}

func (a *Action) GetOutput() any {
	if a.output != nil {
		return *a.output
	}
	return nil
}

// Improved init with better error handling and flexibility.
func (a *Action) init() (*Action, error) {
	fields := strings.Split(a.Name, ".")
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid action name: %s", a.Name)
	}

	a.Method = fields[len(fields)-1]
	a.Handler = strings.ToLower(a.Name)

	return a, nil
}
