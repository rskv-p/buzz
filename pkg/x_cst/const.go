// file: buzz/pkg/x_cst/const.go

package x_cst

import (
	"fmt"
	"os"
	"path/filepath"
)

type ModuleStatus string

const (
	StatusRunning ModuleStatus = "running"
	StatusDown    ModuleStatus = "down"
)

//---------------------
// Log Levels
//---------------------

// Level represents the log level.
type Level int

// Log level constants.
const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

//---------------------
// Default Configuration Values
//---------------------

// Default values for the logger configuration.
const (
	DefaultLogName    = "MyLogger" // Default logger name
	DefaultLogLevel   = DebugLevel // Corresponding to DebugLevel
	DefaultWithTrace  = false      // Default for WithTrace flag
	DefaultWithCaller = true       // Default for WithCaller flag
)

//---------------------
// Color Constants
//---------------------

// Color constants for styling logs
const (
	ColorBlue40   = "#78a9ff"
	ColorGreen40  = "#42be65"
	ColorGray60   = "#8d8d8d"
	ColorOrange40 = "#ff832b"
	ColorRed60    = "#da1e28"
)

//---------------------
// Log Format Constants
//---------------------

// Default log date format
const DefaultLogDateFormat = "01-02 15:04:05"

// Default separator used in logs
const DefaultLogSeparator = " "

const (
	DefaultAuthEnabled   = false
	DefaultAdminLogin    = "admin"
	DefaultAdminPassword = "admin"
)

//---------------------
// Get Golang Project Root
//---------------------

// GetGolangProjectRoot searches for the root of the Go project where the go.mod file is located.
func GetGolangProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for the go.mod file in the current directory or its parent directories
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil // Project root found
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break // Reached the root of the file system
		}
		dir = parentDir
	}

	return "", fmt.Errorf("go.mod not found")
}

//---------------------
// Get User's Home Directory
//---------------------

// GetUserHomeDir returns the user's home directory.
func GetUserHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	return homeDir, nil
}

//---------------------
// Default Paths
//---------------------

// DefaultWorkspacePath — the default workspace path.
var DefaultWorkspacePath string

// Paths to subdirectories within the workspace.
var (
	LogFolderPath  string
	DataFolderPath string
	BinFolderPath  string
	KeysFolderPath string
)

// DefaultLogFilePath — path to the log file within the workspace.
var DefaultLogFilePath string
var DefaultAuthDBPath string

//---------------------
// Initialization of Paths
//---------------------

// init initializes the paths based on the project root or user home directory.
func Init() {
	var err error

	// Get the project root or user's home directory
	projectRoot, err := GetGolangProjectRoot()
	if err != nil {
		// If project is not found, try using the user's home directory
		projectRoot, err = GetUserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("Failed to get project root or user home directory: %v", err))
		}
	}

	// Set the default workspace path
	DefaultWorkspacePath = filepath.Join(projectRoot, ".space")

	// Set paths for subdirectories
	LogFolderPath = filepath.Join(DefaultWorkspacePath, "log")
	DataFolderPath = filepath.Join(DefaultWorkspacePath, "data")
	BinFolderPath = filepath.Join(DefaultWorkspacePath, "bin")
	KeysFolderPath = filepath.Join(DefaultWorkspacePath, "keys")

	// Set path to the log file within the log folder
	DefaultLogFilePath = filepath.Join(LogFolderPath, "buzz.log")
	DefaultAuthDBPath = filepath.Join(DataFolderPath, "auth.db")
	CreateWorkspaceStructure()
}

//---------------------
// Create Workspace Structure
//---------------------

// CreateWorkspaceStructure checks the existence of all necessary directories and creates them if they do not exist.
func CreateWorkspaceStructure() error {
	// List of paths to be created
	directories := []string{
		DefaultWorkspacePath, // .space
		LogFolderPath,        // .space/log
		DataFolderPath,       // .space/data
		BinFolderPath,        // .space/bin
		KeysFolderPath,       // .space/keys
	}

	// Check and create all directories with 755 permissions if they do not exist
	for _, dir := range directories {
		// Check if the directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// If the directory does not exist, create it
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		} else if err != nil {
			// If there was another error during the check, return an error
			return fmt.Errorf("error checking directory %s: %v", dir, err)
		}
	}

	// Check and create the log file with 644 permissions
	// Create the log file only if it does not exist
	file, err := os.OpenFile(DefaultLogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file %s: %v", DefaultLogFilePath, err)
	}
	defer file.Close()

	return nil
}
