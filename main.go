package main

import (
	"log"

	"github.com/rskv-p/buzz/bus/bus_auth"
	"github.com/rskv-p/buzz/bus/bus_client"
	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_core"
	"github.com/rskv-p/buzz/bus/bus_type"
)

func main() {
	// Load configuration
	cfg, err := bus_comm.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the bus system
	bus := bus_core.NewBus(cfg.BusConfig.Name, nil, 5)
	bus_comm.Infof("Bus system initialized with name: %s", cfg.BusConfig.Name)

	// Initialize default admin (if authentication is enabled)
	if cfg.AuthConfig.AuthEnabled {
		if err := bus_auth.InitDB(); err != nil {
			log.Fatalf("Failed to initialize database for authentication: %v", err)
		}
		bus_comm.Infof("Database initialized and default admin user created")
	}

	// Example client setup
	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)

	// Add the client to the bus
	if err := bus.AddClient(client, 1); err != nil {
		log.Fatalf("Failed to add client: %v", err)
	}
	bus_comm.Infof("Client added with ID 1")

	// Subscribe the client to a subject
	if err := client.Subscribe("test.topic", "testQueue"); err != nil {
		log.Fatalf("Failed to subscribe client to topic: %v", err)
	}
	bus_comm.Infof("Client subscribed to 'test.topic' with queue 'testQueue'")

	// Start processing messages
	bus.ProcessMessages()

	// Simulate message publishing
	message := []byte("Test message for bus")
	if err := bus.Publish([]byte("test.topic"), []byte("testQueue"), message); err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	bus_comm.Infof("Message published to 'test.topic' with queue 'testQueue'")

	// Keep the main goroutine running
	select {}
}
