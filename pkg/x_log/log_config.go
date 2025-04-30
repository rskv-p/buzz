// file: buzz/pkg/x_log/log_config.go

package x_log

import (
	"github.com/rskv-p/buzz/pkg/x_cst"
)

var configLogger *Config

//---------------------
// Logger Configuration
//---------------------

// LoggerConfig holds the configuration for the logger.
type Config struct {
	Name           string `json:"name"`           // Logger name
	Level          int    `json:"level"`          // Log level (e.g., Debug, Info)
	OutputFilePath string `json:"outputFilePath"` // Path to the log output file
	WithTrace      bool   `json:"withTrace"`      // Whether to include trace info
	WithCaller     bool   `json:"withCaller"`     // Whether to include caller info
}

//---------------------
// Load Configuration
//---------------------

// LoadConfig loads the configuration from a file, falling back to defaults if necessary.
func LoadConfig() (*Config, error) {

	return &Config{
		Name:           x_cst.DefaultLogName,
		Level:          int(x_cst.DefaultLogLevel), // Cast to int
		OutputFilePath: x_cst.DefaultLogFilePath,
		WithTrace:      x_cst.DefaultWithTrace,
		WithCaller:     x_cst.DefaultWithCaller,
	}, nil
}
