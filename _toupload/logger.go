// file: buzz/pkg/x_log/logger.go

package x_log

import (
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/rskv-p/buzz/pkg/x_cst"
)

//---------------------
// TYPES
//---------------------

// Logger interface defines methods for logging at various levels and formats.
type Logger interface {
	LogEnabled(level x_cst.Level) bool
	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
}

//---------------------
// LOG LEVELS
//---------------------

const (
	TraceLevel x_cst.Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

//---------------------
// ENVIRONMENT
//---------------------

const (
	DefaultLogDateFormat = "01-02 15:04:05"
	DefaultLogSeparator  = " "
)

//---------------------
// GLOBALS
//---------------------

var (
	rootLogger Logger
)

// ---------------------
// INITIALIZATION
// ---------------------

func Init() {
	configLogger, _ = LoadConfig()

	// Initialize the logger based on the config
	rootLogger = createZapRootLogger(configLogger)

}

//---------------------
// ACCESSORS
//---------------------

func RootLogger() Logger {
	return rootLogger
}

func Sync() {
	syncLogger(rootLogger)
}

//---------------------
// DEFAULT LOGGER SHORTCUTS
//---------------------

func Trace(args ...any) { rootLogger.Trace(args...) }
func Debug(args ...any) { rootLogger.Debug(args...) }
func Info(args ...any)  { rootLogger.Info(args...) }
func Warn(args ...any)  { rootLogger.Warn(args...) }
func Error(args ...any) { rootLogger.Error(args...) }

//---------------------
// Constants - Color Definitions
//---------------------

const (
	ColorBlue40   = "#78a9ff"
	ColorGreen40  = "#42be65"
	ColorGray60   = "#8d8d8d"
	ColorOrange40 = "#ff832b"
	ColorRed60    = "#da1e28"
)

//---------------------
// TYPES - LOGGER STYLING
//---------------------

// Styles represents the styles used in the logger for various log levels, keys, and values.
type Styles struct {
	Out               io.Writer                 // Reserved output writer
	Timestamp         lipgloss.Style            // Timestamp style
	Levels            map[string]lipgloss.Style // Level-specific styles
	Keys              map[string]lipgloss.Style // Field key styles
	Values            map[string]lipgloss.Style // Field value styles
	DefaultKeyStyle   lipgloss.Style            // Fallback for unknown keys
	DefaultValueStyle lipgloss.Style            // Fallback for unknown values
}

//---------------------
// DEFAULT STYLES (DARK THEME)
//---------------------

// syncLogger flushes the logger's buffered output.
func syncLogger(logger Logger) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		_ = impl.mainLogger.Sync()
	}
}
