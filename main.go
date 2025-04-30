// file: buzz/main.go

package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/rskv-p/buzz/mod/m_bus"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/rskv-p/buzz/pkg/x_log"
)

//-----------------------------------------
//  Main Entry Point
//-----------------------------------------

func main() {
	//-----------------------------------------
	//  Initialization
	//-----------------------------------------

	x_init.Init() // Initialize config, logging, etc.

	//-----------------------------------------
	//  Resolve Host and Port
	//-----------------------------------------

	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1" // Default to IPv4
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080" // Default port
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		x_log.Error("Invalid port number:", portStr)
		return
	}

	//-----------------------------------------
	//  Create Bus Module
	//-----------------------------------------

	busModule := m_bus.NewBusModule("bus", "secretKey", 10, host, port)
	m_bus.RegisterPublicActions(&busModule.Module)

	//-----------------------------------------
	//  Start HTTP Server
	//-----------------------------------------

	go func() {
		if err := http.ListenAndServe(":"+portStr, nil); err != nil {
			x_log.Error("Failed to start HTTP server:", err)
		}
	}()

	x_log.Info("Server started on port", portStr)

	//-----------------------------------------
	//  Block Forever
	//-----------------------------------------

	select {}
}
