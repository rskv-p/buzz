// file: buzz/pkg/x_log/zap.go

package x_log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rskv-p/buzz/pkg/x_cst"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//---------------------
// ZAP LOGGER IMPL
//---------------------

// zapLoggerImpl implements the Logger interface using the Zap logger.
type zapLoggerImpl struct {
	loggerLevel *zap.AtomicLevel
	mainLogger  *zap.SugaredLogger
	traceLogger *zap.SugaredLogger
}

// DebugEnabled checks if the debug level logging is enabled.
func (l *zapLoggerImpl) LogEnabled(level x_cst.Level) bool {
	return l.loggerLevel.Enabled(toZapLogLevel(level))
}

// Trace logs messages at the trace level.
func (l *zapLoggerImpl) Trace(args ...any) {
	if l.LogEnabled(TraceLevel) && l.traceLogger != nil {
		l.traceLogger.Debug(args...)
	}
}

// Tracef logs formatted trace messages.
func (l *zapLoggerImpl) Tracef(format string, args ...any) {
	if l.LogEnabled(TraceLevel) && l.traceLogger != nil {
		l.traceLogger.Debugf(format, args...)
	}
}

// Debug logs messages at the debug level.
func (l *zapLoggerImpl) Debug(args ...any) { l.mainLogger.Debug(args...) }

// Info logs messages at the info level.
func (l *zapLoggerImpl) Info(args ...any) { l.mainLogger.Info(args...) }

// Warn logs messages at the warn level.
func (l *zapLoggerImpl) Warn(args ...any) { l.mainLogger.Warn(args...) }

// Error logs messages at the error level.
func (l *zapLoggerImpl) Error(args ...any) { l.mainLogger.Error(args...) }

// ---------------------
// ROOT LOGGER FROM CONFIG
// ---------------------
func createZapRootLogger(configLogger *Config) Logger {
	// Create the console encoder for pretty output
	//consoleEncoder := zapcore.NewConsoleEncoder(newConsoleEncoderConfig())

	// Create the JSON encoder for file output
	jsonEncoder := zapcore.NewJSONEncoder(newJSONEncoderConfig())

	// Create the output writer (both console and file)
	writer := createLogOutput()

	// Set the logging level
	atomicLevel := zap.NewAtomicLevelAt(toZapLogLevel(x_cst.DefaultLogLevel))

	// Create the core with both console and file output
	core := zapcore.NewCore(
		jsonEncoder, // JSON encoder for the file
		writer,      // Write to console and file
		atomicLevel,
	)

	// Add caller information (file and line number)
	opts := []zap.Option{zap.AddCallerSkip(1)} // Skip one stack frame for proper location
	if configLogger.WithCaller {
		opts = append(opts, zap.AddCaller()) // Adds filename and line number
	}

	// Create the logger
	zl := zap.New(core, opts...)
	var traceLogger *zap.SugaredLogger
	if configLogger.WithTrace {
		traceLogger = zl.Sugar()
	}

	return &zapLoggerImpl{
		loggerLevel: &atomicLevel,
		mainLogger:  zl.Named(configLogger.Name).Sugar(),
		traceLogger: traceLogger,
	}
}

//---------------------
// ENCODER CONFIG
//---------------------

// newJSONEncoderConfig creates a new encoder configuration for the logger (file).
func newJSONEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.ConsoleSeparator = getLogSeparatorOrDefault()

	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("%v", t))
	}

	cfg.EncodeName = func(name string, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(name)
	}

	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.CallerKey = "caller"
	return cfg
}

// createLogOutput creates the output writer based on config.
func createLogOutput() zapcore.WriteSyncer {
	// Always write to the console (stdout)
	consoleWriter := zapcore.Lock(os.Stdout) // Writing to the console (stdout)

	logFilePath := configLogger.OutputFilePath

	// Ensure that the directory exists before writing to the log file
	dir := filepath.Dir(logFilePath)

	// Check if the directory exists, if not, create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("Failed to create log directory: %v", err))
		}
	}

	// Open the log file (create if it doesn't exist, append if it does)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file: %v", err))
	}

	// Combine both console and file outputs (multiple outputs)
	// Write to both file and console
	return zapcore.NewMultiWriteSyncer(consoleWriter, zapcore.AddSync(file))
}

// getLogSeparatorOrDefault retrieves the log separator from the environment or returns default.
func getLogSeparatorOrDefault() string {
	// Default separator is " " if it's not specified in the config or environment.
	return " "
}

// toZapLogLevel converts a Level type to zapcore.Level.
func toZapLogLevel(level x_cst.Level) zapcore.Level {
	switch level {
	case DebugLevel, TraceLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
