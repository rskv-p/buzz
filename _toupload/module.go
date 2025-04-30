// file: buzz/mod/module.go

package mod

import (
	"errors"

	"github.com/rskv-p/buzz/typ"
)

//---------------------
// Types
//---------------------

// Module is a reusable component that can register actions.
type Module struct {
	Name    string
	Actions []typ.ActionDef

	OnInit func() error
	OnStop func() error

	actionHandlers map[string]typ.ActionHandler
}

//---------------------
// Lifecycle
//---------------------

// Name returns the module name.
func (m *Module) GetName() string {
	return m.Name
}

// Init initializes the module and registers actions.
func (m *Module) Init() error {
	// Register actions
	for _, a := range m.Actions {
		if err := m.RegisterAction(a); err != nil {
			return err
		}
	}

	// Initialize actionHandlers map if not initialized
	if m.actionHandlers == nil {
		m.actionHandlers = make(map[string]typ.ActionHandler)
	}

	// Run custom init (if any)
	if m.OnInit != nil {
		return m.OnInit()
	}

	return nil
}

// Stop stops the module and executes custom stop logic (if any).
func (m *Module) Stop() error {
	// Run custom stop (if any)
	if m.OnStop != nil {
		return m.OnStop()
	}

	return nil
}

//---------------------
// Actions Management
//---------------------

// RegisterAction registers a new action. Returns an error if already exists.
func (m *Module) RegisterAction(action typ.ActionDef) error {
	for _, a := range m.Actions {
		if a.Name == action.Name {
			return errors.New("action already registered: " + action.Name)
		}
	}
	m.Actions = append(m.Actions, action)
	return nil
}

// UnregisterAction removes an action by name. Returns an error if not found.
func (m *Module) UnregisterAction(name string) error {
	for i, a := range m.Actions {
		if a.Name == name {
			m.Actions = append(m.Actions[:i], m.Actions[i+1:]...)
			return nil
		}
	}
	return errors.New("action not found: " + name)
}

// GetActions returns all registered actions.
func (m *Module) GetActions() []typ.ActionDef {
	return m.Actions
}

// GetAction retrieves an action by name. Returns the action and a boolean flag if found.
func (m *Module) GetAction(name string) (typ.ActionDef, bool) {
	for _, a := range m.Actions {
		if a.Name == name {
			return a, true
		}
	}
	return typ.ActionDef{}, false
}

//---------------------
// Action Handlers Management
//---------------------

// RegisterHandler registers a new handler. Returns an error if already exists.
func (m *Module) RegisterHandler(name string, handler typ.ActionHandler) error {
	if m.actionHandlers == nil {
		m.actionHandlers = make(map[string]typ.ActionHandler)
	}
	if _, exists := m.actionHandlers[name]; exists {
		return errors.New("handler already registered: " + name)
	}
	m.actionHandlers[name] = handler
	return nil
}

// UnregisterHandler removes a handler by name. Returns an error if not found.
func (m *Module) UnregisterHandler(name string) error {
	if m.actionHandlers == nil {
		return errors.New("no handlers registered")
	}
	if _, exists := m.actionHandlers[name]; !exists {
		return errors.New("handler not found: " + name)
	}
	delete(m.actionHandlers, name)
	return nil
}

// GetHandler retrieves a handler by name. Returns the handler and a boolean flag if found.
func (m *Module) GetHandler(name string) (typ.ActionHandler, bool) {
	handler, ok := m.actionHandlers[name]
	return handler, ok
}

// GetHandlers returns all registered handlers.
func (m *Module) GetHandlers() map[string]typ.ActionHandler {
	if m.actionHandlers == nil {
		return map[string]typ.ActionHandler{}
	}
	return m.actionHandlers
}
