// file: buz/bus/bus_comm/config.go

package bus_comm

import (
	"encoding/json"
	"fmt"
	"os"
)

//-------------------------------------------------
// Config Structs - Configuration structures for the bus system
//-------------------------------------------------

// Config holds the overall configuration for the bus system including log, bus, and authentication settings.
type Config struct {
	LoggerConfig LoggerConfig `json:"loggerConfig"` // Configuration for logging
	BusConfig    BusConfig    `json:"busConfig"`    // Configuration for the bus
	AuthConfig   AuthConfig   `json:"authConfig"`   // Configuration for authentication
}

// LoggerConfig holds the configuration settings for logging.
type LoggerConfig struct {
	Name           string `json:"name"`           // Logger name
	Level          Level  `json:"level"`          // Logging level
	Format         Format `json:"format"`         // Log format (console or json)
	OutputFilePath string `json:"outputFilePath"` // Log outputs (stdout, stderr, file)
	WithTrace      bool   `json:"withTrace"`      // Enable TRACE level logging
	WithCaller     bool   `json:"withCaller"`     // Include caller information in logs
	RotateMaxMB    int    `json:"rotateMaxMB"`    // Max log file size in MB
	RotateMaxAge   int    `json:"rotateMaxAge"`   // Max log retention period in days
	RotateBackups  int    `json:"rotateBackups"`  // Number of old log files to retain
	RotateCompress bool   `json:"rotateCompress"` // Compress old log files
}

// BusConfig holds configuration for the message bus system.
type BusConfig struct {
	Version    string `json:"version"`    // Bus version
	Name       string `json:"name"`       // Bus name
	Repository string `json:"repository"` // Repository for the bus system
}

// AuthConfig holds configuration for authentication settings.
type AuthConfig struct {
	AuthEnabled          bool   `json:"authEnabled"`          // Flag to enable authentication
	DefaultAdminLogin    string `json:"defaultAdminLogin"`    // Default admin login
	DefaultAdminPassword string `json:"defaultAdminPassword"` // Default admin password
	DBpath               string `json:"dbPath"`               // Path to the database for authentication
}

//-------------------------------------------------
// LoadConfig - Loads configuration from the default.json file
//-------------------------------------------------

// LoadConfig loads the configuration from the default.json file
func LoadConfig() (*Config, error) {
	var config Config

	// Open the default.json configuration file
	file, err := os.Open(".bus/cfg/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	// Decode the JSON content from the file into the config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	// Return the configuration
	return &config, nil
}
