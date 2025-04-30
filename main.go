package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/rskv-p/buzz/mod/m_bus"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/rskv-p/buzz/pkg/x_log"
)

func main() {
	// Initialize the application (config, logging, etc.)
	x_init.Init()

	// Define the host and port
	host := os.Getenv("HOST") // Fetch from environment or use default
	if host == "" {
		host = "127.0.0.1" // Default to IPv4 if not provided
	}

	portStr := os.Getenv("PORT") // Fetch from environment or use default
	if portStr == "" {
		portStr = "8080" // Default port
	}

	// Convert port from string to int
	port, err := strconv.Atoi(portStr)
	if err != nil {
		x_log.Error("Invalid port number:", portStr)
		return
	}

	// Create the bus module (supports both IPv4 and IPv6)
	busModule := m_bus.NewBusModule("bus", "secretKey", 10, host, port)

	// After creating the module, register public actions
	m_bus.RegisterPublicActions(&busModule.Module)

	// Start the HTTP server
	go func() {
		// Listen and serve on the specified port
		if err := http.ListenAndServe(":"+portStr, nil); err != nil {
			x_log.Error("Failed to start HTTP server:", err)
		}
	}()

	x_log.Info("Server started on port", portStr)
	select {} // Block forever to keep the server running
}
