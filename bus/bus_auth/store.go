// file: buz/bus/bus_auth/store.go

package bus_auth

import (
	"fmt"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//---------------------
// GLOBAL VARIABLES
//---------------------

var db *gorm.DB

//---------------------
// STRUCTS: USER, PERMISSION, GROUP, SERVICE ACCOUNT
//---------------------

// User represents a user in the system with permissions and roles.
type User struct {
	ID          uint         `gorm:"primaryKey"`
	Username    string       `gorm:"uniqueIndex"` // Unique username for the user
	Password    string       // Hashed password for authentication
	Permissions []Permission `gorm:"many2many:user_permissions;"` // Many-to-many relationship with permissions
	Groups      []Group      `gorm:"many2many:user_groups;"`      // Many-to-many relationship with groups
}

// Permission represents a topic-based permission.
type Permission struct {
	ID   uint   `gorm:"primaryKey"`  // Unique ID for the permission
	Name string `gorm:"uniqueIndex"` // Unique name for the permission
}

// Group represents a group of users with associated permissions.
type Group struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"uniqueIndex"`                  // Unique group name
	Permissions []Permission `gorm:"many2many:group_permissions;"` // Many-to-many relationship with permissions
}

type UserPermission struct {
	UserID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

// ServiceAccount represents a service account used for service-to-service authentication.
type ServiceAccount struct {
	ID          uint         `gorm:"primaryKey"`
	ServiceName string       `gorm:"uniqueIndex"` // Unique name for the service
	APIKey      string       // Secret API key for the service account
	Permissions []Permission `gorm:"many2many:user_permissions;"` // Many-to-many relationship with permissions
	Groups      []Group      `gorm:"many2many:user_groups;"`      // Many-to-many relationship with groups
}

//---------------------
// DATABASE INITIALIZATION
//---------------------

// InitDB initializes the database connection and ensures migrations are applied.
func InitDB() error {
	// Load configuration from bus_comm
	//cfg := bus_comm.Cfg.AuthConfig
	// Initialize SQLite database connection
	var err error
	db, err = gorm.Open(sqlite.Open(bus_comm.Cfg.AuthConfig.DBpath), &gorm.Config{})
	if err != nil {
		bus_comm.Errorf("Failed to connect to database at %s: %v", bus_comm.Cfg.AuthConfig.DBpath, err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	bus_comm.Infof("Successfully connected to the SQLite database at %s", bus_comm.Cfg.AuthConfig.DBpath)

	// Migrate the schema
	err = db.AutoMigrate(&User{}, &Permission{}, &Group{}, &ServiceAccount{}, &UserPermission{})
	if err != nil {
		bus_comm.Errorf("Failed to migrate database schema: %v", err)
		return fmt.Errorf("failed to migrate database schema: %v", err)
	}
	bus_comm.Infof("Database schema migration completed successfully")

	// Ensure the default admin user is created if it doesn't exist
	return createDefaultAdmin()
}

//---------------------
// ADMIN USER CREATION
//---------------------

// createDefaultAdmin ensures that the default admin user is created if not already present.
func createDefaultAdmin() error {
	var admin User
	// Check if the default admin exists or create it
	err := db.Where("username = ?", bus_comm.Cfg.AuthConfig.DefaultAdminLogin).First(&admin).Error
	if err == nil {
		// Admin user already exists, skip creation
		bus_comm.Infof("Admin user %s already exists. Skipping creation.", bus_comm.Cfg.AuthConfig.DefaultAdminLogin)
		return nil
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(bus_comm.Cfg.AuthConfig.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		bus_comm.Errorf("Failed to hash password for admin user %s: %v", bus_comm.Cfg.AuthConfig.DefaultAdminLogin, err)
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Create default admin user
	admin = User{
		Username: bus_comm.Cfg.AuthConfig.DefaultAdminLogin,
		Password: string(hashedPassword),
	}

	// Create the admin permission (full access)
	adminPermission := Permission{Name: bus_comm.Cfg.AuthConfig.DefaultAdminLogin}
	if err := db.Create(&adminPermission).Error; err != nil {
		bus_comm.Errorf("Failed to create admin permission: %v", err)
		return fmt.Errorf("failed to create admin permission: %v", err)
	}

	// Save the user and assign permission
	admin.Permissions = append(admin.Permissions, adminPermission)
	if err := db.Create(&admin).Error; err != nil {
		bus_comm.Errorf("Failed to create admin user %s: %v", bus_comm.Cfg.AuthConfig.DefaultAdminLogin, err)
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	bus_comm.Infof("Successfully created default admin user %s", bus_comm.Cfg.AuthConfig.DefaultAdminLogin)

	return nil
}
