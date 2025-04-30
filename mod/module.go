package mod

import (
	"errors"

	"github.com/rskv-p/buzz/typ"
)

var _ typ.IModule = (*Module)(nil) // Ensure RequestHeaders implements IRequestHeaders

//---------------------
// Types
//---------------------

// Module is a reusable component that can register actions.
type Module struct {
	Name               string
	Version            string `json:"version"`
	Description        string `json:"description"`
	QueueGroup         string `json:"queue_group"`
	QueueGroupDisabled bool   `json:"queue_group_disabled"`

	Actions map[string]typ.IAction
}

//---------------------
// Lifecycle
//---------------------

// NewModule creates a new Module with default settings.
func NewModule(name string) *Module {
	return &Module{
		Name:    name,
		Actions: make(map[string]typ.IAction),
	}
}

// GetName returns the module name.
func (m *Module) GetName() string {
	return m.Name
}

// Init initializes the module and registers actions.
func (m *Module) Init() error {
	for _, a := range m.Actions {
		if err := m.RegisterAction(a); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the module (can be overridden if embedded).
func (m *Module) Start() error {
	return nil
}

// Stop stops the module (can be overridden if embedded).
func (m *Module) Stop() error {
	return nil
}

//---------------------
// Actions Management
//---------------------

// RegisterAction registers a new action. Returns an error if already exists.
func (m *Module) RegisterAction(action typ.IAction) error {
	if _, exists := m.Actions[action.GetName()]; exists {
		return errors.New("action already registered: " + action.GetName())
	}
	m.Actions[action.GetName()] = action
	return nil
}

// UnregisterAction removes an action by name. Returns an error if not found.
func (m *Module) UnregisterAction(name string) error {
	if _, exists := m.Actions[name]; !exists {
		return errors.New("action not found: " + name)
	}
	delete(m.Actions, name)
	return nil
}

// GetActions returns all registered actions.
func (m *Module) GetActions() []typ.IAction {
	actions := make([]typ.IAction, 0, len(m.Actions))
	for _, action := range m.Actions {
		actions = append(actions, action)
	}
	return actions
}

// GetAction retrieves an action by name. Returns the action and a boolean flag if found.
func (m *Module) GetAction(name string) (typ.IAction, bool) {
	action, exists := m.Actions[name]
	return action, exists
}
