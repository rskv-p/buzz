// file: buzz/srv/service.go

package srv

import (
	"log"

	"github.com/rskv-p/buzz/typ"
)

var _ typ.IService = (*Service)(nil)

//-----------------------------------------
//  Service Struct
//-----------------------------------------

// Service implements IService with lifecycle management.
type Service struct {
	Name      string         `json:"name"` // Service name
	Modules   []typ.IModule  // Registered modules
	busClient typ.IBusClient // Connected bus client
}

//-----------------------------------------
//  NewService
//-----------------------------------------

// NewService creates a new service instance with modules.
func NewService(bus typ.IBusClient, name string, modules ...typ.IModule) *Service {
	svc := &Service{
		Name:      name,
		busClient: bus,
		Modules:   modules,
	}

	log.Printf("Service %s created and initialized with %d module(s)", svc.GetName(), len(modules))
	return svc
}

//-----------------------------------------
//  Start
//-----------------------------------------

// Start launches the service and all registered modules.
func (s *Service) Start() error {
	log.Printf("Starting service %s...", s.GetName())

	for _, module := range s.Modules {
		log.Printf("Starting module %s...", module.GetName())

		if err := module.Start(); err != nil {
			log.Printf("Failed to start module %s: %v", module.GetName(), err)
			return err
		}

		log.Printf("Module %s started successfully", module.GetName())
	}

	log.Printf("Service %s started successfully", s.GetName())
	return nil
}

//-----------------------------------------
//  Stop
//-----------------------------------------

// Stop gracefully shuts down the service and all modules.
func (s *Service) Stop() {
	log.Printf("Stopping service %s...", s.GetName())

	for _, module := range s.Modules {
		log.Printf("Stopping module %s...", module.GetName())

		if err := module.Stop(); err != nil {
			log.Printf("Failed to stop module %s: %v", module.GetName(), err)
		} else {
			log.Printf("Module %s stopped successfully", module.GetName())
		}
	}

	log.Printf("Service %s stopped", s.GetName())
}

//-----------------------------------------
//  GetName
//-----------------------------------------

// GetName returns the name of the service.
func (s *Service) GetName() string {
	return s.Name
}
