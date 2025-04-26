// file: buz/bus/bus_comm/zap.go

package bus_comm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var traceLogger *zap.SugaredLogger

//---------------------
// ZAP LOGGER IMPL
//---------------------

// zapLoggerImpl implements the Logger interface using the Zap logger.
type zapLoggerImpl struct {
	loggerLevel *zap.AtomicLevel
	mainLogger  *zap.SugaredLogger
}

// DebugEnabled checks if the debug level logging is enabled.
func (l *zapLoggerImpl) DebugEnabled() bool {
	return l.loggerLevel.Enabled(zapcore.DebugLevel)
}

// TraceEnabled checks if the trace level logging is enabled.
func (l *zapLoggerImpl) TraceEnabled() bool {
	return traceEnabled && l.DebugEnabled()
}

// Trace logs messages at the trace level.
func (l *zapLoggerImpl) Trace(args ...any) {
	if traceEnabled {
		traceLogger.Debug(args...)
	}
}

// Tracef logs formatted trace messages.
func (l *zapLoggerImpl) Tracef(format string, args ...any) {
	if traceEnabled {
		traceLogger.Debugf(format, args...)
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

// Debugf logs formatted debug messages.
func (l *zapLoggerImpl) Debugf(f string, a ...any) { l.mainLogger.Debugf(f, a...) }

// Infof logs formatted info messages.
func (l *zapLoggerImpl) Infof(f string, a ...any) { l.mainLogger.Infof(f, a...) }

// Warnf logs formatted warn messages.
func (l *zapLoggerImpl) Warnf(f string, a ...any) { l.mainLogger.Warnf(f, a...) }

// Errorf logs formatted error messages.
func (l *zapLoggerImpl) Errorf(f string, a ...any) { l.mainLogger.Errorf(f, a...) }

// Structured returns a structured logger instance.
func (l *zapLoggerImpl) Structured() StructuredLogger {
	return &zapStructuredLoggerImpl{lvl: l.loggerLevel, zl: l.mainLogger.Desugar()}
}

//---------------------
// STRUCTURED LOGGER
//---------------------

// zapStructuredLoggerImpl implements the StructuredLogger interface for structured logging.
type zapStructuredLoggerImpl struct {
	lvl *zap.AtomicLevel
	zl  *zap.Logger
}

// Trace logs messages at the trace level for structured logging.
func (l *zapStructuredLoggerImpl) Trace(msg string, fields ...Field) {
	if traceEnabled {
		l.zl.Debug(msg, convertFields(fields)...) // trace -> debug
	}
}

// Debug logs messages at the debug level for structured logging.
func (l *zapStructuredLoggerImpl) Debug(msg string, fields ...Field) {
	l.zl.Debug(msg, convertFields(fields)...)
}

// Info logs messages at the info level for structured logging.
func (l *zapStructuredLoggerImpl) Info(msg string, fields ...Field) {
	l.zl.Info(msg, convertFields(fields)...)
}

// Warn logs messages at the warn level for structured logging.
func (l *zapStructuredLoggerImpl) Warn(msg string, fields ...Field) {
	l.zl.Warn(msg, convertFields(fields)...)
}

// Error logs messages at the error level for structured logging.
func (l *zapStructuredLoggerImpl) Error(msg string, fields ...Field) {
	l.zl.Error(msg, convertFields(fields)...)
}

// convertFields converts the fields to zap.Field format.
func convertFields(fields []Field) []zap.Field {
	zFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zFields[i] = f.(zap.Field)
	}
	return zFields
}

//---------------------
// ROOT LOGGER FROM CONFIG
//---------------------

// newZapRootLoggerWithOutput creates the root logger using the provided configuration.
func newZapRootLoggerWithOutput(cfg *LoggerConfig) Logger {
	encoderCfg := newEncoderConfig()
	writer := buildZapOutputWithRotation(cfg)

	atomicLevel := zap.NewAtomicLevelAt(toZapLogLevel(cfg.Level))

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		writer,
		atomicLevel,
	)

	opts := []zap.Option{zap.AddCallerSkip(1)}
	if cfg.WithCaller {
		opts = append(opts, zap.AddCaller())
	}

	zl := zap.New(core, opts...)
	if cfg.WithTrace {
		traceLogger = zl.Sugar()
	}

	return &zapLoggerImpl{
		loggerLevel: &atomicLevel,
		mainLogger:  zl.Named(cfg.Name).Sugar(),
	}
}

//---------------------
// ENCODER CONFIG
//---------------------

// newEncoderConfig creates a new encoder configuration for the logger.
func newEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.ConsoleSeparator = getLogSeparator()

	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		formatted := fmt.Sprintf("[%02d-%02d %02d:%02d:%02d]",
			t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(),
		)
		enc.AppendString(logStyles.Timestamp.Render(formatted))
	}

	cfg.EncodeName = func(name string, enc zapcore.PrimitiveArrayEncoder) {
		styled := logStyles.DefaultKeyStyle.Render("[" + name + "]")
		enc.AppendString(styled + " -")
	}

	cfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		level := strings.ToLower(l.String())
		if style, ok := logStyles.Levels[level]; ok {
			enc.AppendString(style.Render(padLevel(level)))
		} else {
			enc.AppendString(padLevel(level))
		}
	}

	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.CallerKey = "caller"
	return cfg
}

// padLevel adjusts the padding for log level strings.
func padLevel(level string) string {
	switch level {
	case "info":
		return "INF"
	case "warn":
		return "WRN"
	case "error":
		return "ERR"
	case "panic":
		return "PNC"
	case "debug":
		return "DBG"
	default:
		return strings.ToUpper(level)
	}
}

//---------------------
// OUTPUT WRITER
//---------------------

// buildZapOutputWithRotation creates the output writer with log rotation.
func buildZapOutputWithRotation(cfg *LoggerConfig) zapcore.WriteSyncer {
	var writers []zapcore.WriteSyncer

	// Always write to the console (stdout)
	writers = append(writers, zapcore.Lock(os.Stdout)) // Writing to the console (stdout)

	// Get the log file path from the config, default to ".bus/log/log.log" if not set
	logFilePath := cfg.OutputFilePath
	if logFilePath == "" {
		logFilePath = ".bus/log/log.log" // Default fallback path if not provided in the config
	}

	// Ensure that the directory exists before writing to the log file
	dir := filepath.Dir(logFilePath)

	// Check if the directory exists, if not, create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("Failed to create log directory: %v", err))
		}
	}

	// Add the log file output with rotation settings
	writers = append(writers, zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,        // Log file path from the config
		MaxSize:    cfg.RotateMaxMB,    // Maximum size before rotation (in MB)
		MaxBackups: cfg.RotateBackups,  // Number of old log files to keep
		MaxAge:     cfg.RotateMaxAge,   // Days to retain old log files
		Compress:   cfg.RotateCompress, // Whether to compress old log files
	}))

	// Combine both console and file outputs (multiple outputs)
	return zapcore.NewMultiWriteSyncer(writers...)
}

//---------------------
// HELPERS
//---------------------

// toZapLogLevel converts a Level type to zapcore.Level.
func toZapLogLevel(level Level) zapcore.Level {
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

// setZapLogLevel sets the log level for the logger.
func setZapLogLevel(logger Logger, level Level) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		impl.loggerLevel.SetLevel(toZapLogLevel(level))
	}
}

// newZapChildLogger creates a child logger with a specified name.
func newZapChildLogger(logger Logger, name string) (Logger, error) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		newZl := impl.mainLogger.Named(name)
		return &zapLoggerImpl{
			loggerLevel: impl.loggerLevel,
			mainLogger:  newZl,
		}, nil
	}
	return nil, fmt.Errorf("invalid zapLoggerImpl")
}

// newZapChildLoggerWithFields creates a child logger with specified fields.
func newZapChildLoggerWithFields(logger Logger, fields ...Field) (Logger, error) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		newZl := impl.mainLogger.With(fields...)
		return &zapLoggerImpl{
			loggerLevel: impl.loggerLevel,
			mainLogger:  newZl,
		}, nil
	}
	return nil, fmt.Errorf("invalid zapLoggerImpl")
}

// zapSync flushes the logger's buffered output.
func zapSync(logger Logger) {
	if impl, ok := logger.(*zapLoggerImpl); ok {
		_ = impl.mainLogger.Sync()
	}
}
