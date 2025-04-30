// file:buzz/mod/m_bus/bus_auth/config.go

package bus_auth

import (
	"github.com/rskv-p/buzz/pkg/x_cst"
)

// Configuration holds the entire configuration for the bus system
var configAuth *Config

//---------------------
// Config Structs - Configuration structures for the bus system
//---------------------

// AuthConfig holds configuration for authentication settings.
type Config struct {
	AuthEnabled   bool   `json:"authEnabled"`          // Flag to enable authentication
	AdminLogin    string `json:"defaultAdminLogin"`    // Default admin login
	AdminPassword string `json:"defaultAdminPassword"` // Default admin password
	DBpath        string `json:"dbPath"`               // Path to the database for authentication
}

func Init() {
	configAuth, _ = LoadConfig()
}

//---------------------
// Load Configuration
//---------------------

// LoadConfig loads the configuration from a file, falling back to defaults if necessary.
func LoadConfig() (*Config, error) {
	return &Config{
		AuthEnabled:   x_cst.DefaultAuthEnabled,
		AdminLogin:    x_cst.DefaultAdminLogin,
		AdminPassword: x_cst.DefaultAdminPassword,
		DBpath:        x_cst.DefaultAuthDBPath,
	}, nil
}
