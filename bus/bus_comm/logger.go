// file: buz/bus/bus_comm/logger.go

package bus_comm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//---------------------
// TYPES
//---------------------

// Level represents the log level.
type Level int

// Format represents the log format.
type Format int

// Logger interface defines methods for logging at various levels and formats.
type Logger interface {
	TraceEnabled() bool
	DebugEnabled() bool

	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)

	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)

	Structured() StructuredLogger
}

// StructuredLogger interface defines methods for structured logging.
type StructuredLogger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field represents a field in structured logging.
type Field = any

//---------------------
// LOG LEVELS
//---------------------

// Log levels for the logger.
const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

//---------------------
// LOG FORMATS
//---------------------

// Log formats for the logger.
const (
	FormatConsole Format = iota
	FormatJson
)

//---------------------
// ENVIRONMENT
//---------------------

// Environment variables for logging configuration.
const (
	EnvKeyLogCtx        = "MINI_LOG_CTX"
	EnvKeyLogDateFormat = "MINI_LOG_DTFORMAT"
	EnvKeyLogLevel      = "MINI_LOG_LEVEL"
	EnvKeyLogFormat     = "MINI_LOG_FORMAT"
	EnvKeyLogSeparator  = "MINI_LOG_SEPARATOR"
	EnvLogConsoleStream = "MINI_LOG_CONSOLE_STREAM"
	EnvLogFilePath      = "MINI_LOG_FILE"
	EnvLogFileMaxMB     = "MINI_LOG_FILE_MAX_MB"
	EnvLogFileMaxAge    = "MINI_LOG_FILE_MAX_AGE"
	EnvLogFileMaxBack   = "MINI_LOG_FILE_BACKUPS"
	EnvLogFileCompress  = "MINI_LOG_FILE_COMPRESS"

	DefaultLogDateFormat = "01-02 15:04:05"
	DefaultLogLevel      = DebugLevel
	DefaultLogFormat     = FormatConsole
	DefaultLogSeparator  = " "
)

//---------------------
// GLOBALS
//---------------------

// Global variables for the logger and configuration.
var (
	rootLogger   Logger
	ctxLogging   bool
	traceEnabled = true
	logStyles    = DefaultStylesDark()
)

// ---------------------
// INITIALIZATION
// ---------------------

// Configuration for logging
var Cfg *Config

// init function initializes the logger configuration and settings.
// init function initializes the logger configuration and settings.
func init() {
	var err error
	Cfg, err = LoadConfig()
	if err != nil {
		// Log the error and stop execution if loading the configuration fails
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Ensure the configuration is properly loaded
	if Cfg == nil {
		panic("Configuration is nil after loading.")
	}

	// Initialize the logger
	rootLogger = newZapRootLoggerWithOutput(&Cfg.LoggerConfig)
	SetLogLevel(rootLogger, Cfg.LoggerConfig.Level)
}

// IsDebugEnabled checks if the current log level is Debug or lower.
func IsDebugEnabled() bool {
	return Cfg.LoggerConfig.Level <= DebugLevel
}

//---------------------
// ACCESSORS
//---------------------

// CtxLoggingEnabled checks if context logging is enabled.
func CtxLoggingEnabled() bool {
	return ctxLogging
}

// RootLogger returns the root logger instance.
func RootLogger() Logger {
	return rootLogger
}

// SetLogLevel sets the log level for the logger.
func SetLogLevel(logger Logger, level Level) {
	setZapLogLevel(logger, level)
}

// Sync synchronizes the logger.
func Sync() {
	zapSync(rootLogger)
}

//---------------------
// CHILD LOGGER HELPERS
//---------------------

// ChildLogger creates a child logger with a specified name.
func ChildLogger(logger Logger, name string) Logger {
	child, err := newZapChildLogger(logger, name)
	if err != nil {
		rootLogger.Warnf("unable to create child logger [%s]: %s", name, err)
		return logger
	}
	return child
}

// ChildLoggerWithFields creates a child logger with specified fields.
func ChildLoggerWithFields(logger Logger, fields ...Field) Logger {
	child, err := newZapChildLoggerWithFields(logger, fields...)
	if err != nil {
		rootLogger.Warnf("unable to create child logger with fields: %s", err)
		return logger
	}
	return child
}

// CreateLoggerFromRef creates a logger from a reference string, categorizing by activity, trigger, or connector.
func CreateLoggerFromRef(logger Logger, contributionType, ref string) Logger {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimSuffix(ref, "/")
	dirs := strings.Split(ref, "/")

	switch {
	case len(dirs) >= 3:
		name := dirs[len(dirs)-1]
		acType := dirs[len(dirs)-2]
		if acType == "activity" || acType == "trigger" || acType == "connector" {
			category := dirs[len(dirs)-3]
			return ChildLogger(logger, strings.ToLower(category+"."+acType+"."+name))
		}
		return ChildLogger(logger, strings.ToLower(acType+"."+contributionType+"."+name))
	default:
		return ChildLogger(logger, strings.ToLower(contributionType+"."+filepath.Base(ref)))
	}
}

//---------------------
// UTILITIES
//---------------------

// ToLogLevel converts a string level to a Level type.
func ToLogLevel(levelStr string) Level {
	switch strings.ToUpper(levelStr) {
	case "TRACE":
		return DebugLevel
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		return DefaultLogLevel
	}
}

// getLogSeparator retrieves the log separator from the environment.
func getLogSeparator() string {
	if v, ok := os.LookupEnv(EnvKeyLogSeparator); ok && len(v) > 0 {
		return v
	}
	return DefaultLogSeparator
}

//---------------------
// STYLE HELPERS
//---------------------

// StyledLevel styles the log level for output.
func StyledLevel(level string) string {
	if s, ok := logStyles.Levels[level]; ok {
		return s.Render(strings.ToUpper(level))
	}
	return logStyles.DefaultKeyStyle.Render(strings.ToUpper(level))
}

// StyledKey styles the key for output.
func StyledKey(key string) string {
	if s, ok := logStyles.Keys[key]; ok {
		return s.Render(key)
	}
	return logStyles.DefaultKeyStyle.Render(key)
}

// StyledValue styles the value for output.
func StyledValue(key, value string) string {
	if s, ok := logStyles.Values[key]; ok {
		return s.Render(value)
	}
	return logStyles.DefaultValueStyle.Render(value)
}

// RenderStyledFields renders a map of styled fields.
func RenderStyledFields(fields map[string]string) string {
	var sb strings.Builder
	for k, v := range fields {
		sb.WriteString(StyledKey(k))
		sb.WriteString("=")
		sb.WriteString(StyledValue(k, v))
		sb.WriteString("  ")
	}
	return sb.String()
}

//---------------------
// DEFAULT LOGGER SHORTCUTS
//---------------------

// Shortcut methods for logging at various levels.

func Trace(args ...any) { rootLogger.Trace(args...) }
func Debug(args ...any) { rootLogger.Debug(args...) }
func Info(args ...any)  { rootLogger.Info(args...) }
func Warn(args ...any)  { rootLogger.Warn(args...) }
func Error(args ...any) { rootLogger.Error(args...) }

func Tracef(format string, args ...any) { rootLogger.Tracef(format, args...) }
func Debugf(format string, args ...any) { rootLogger.Debugf(format, args...) }
func Infof(format string, args ...any)  { rootLogger.Infof(format, args...) }
func Warnf(format string, args ...any)  { rootLogger.Warnf(format, args...) }
func Errorf(format string, args ...any) { rootLogger.Errorf(format, args...) }
