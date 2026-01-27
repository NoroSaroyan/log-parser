package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// InitLogger initializes zap logger with the specified level
func InitLogger(level string) error {
	// Determine log level
	logLevel := zapcore.InfoLevel
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = zapcore.DebugLevel
	case "INFO":
		logLevel = zapcore.InfoLevel
	case "WARN":
		logLevel = zapcore.WarnLevel
	case "ERROR":
		logLevel = zapcore.ErrorLevel
	case "FATAL":
		logLevel = zapcore.FatalLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	// Check if running in terminal
	isTerminal := isTerminalOutput()

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if !isTerminal {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	// Create logger with stack trace on error and above
	log = zap.New(core, zap.AddCallerSkip(1))

	return nil
}

// GetLogger returns the logger instance for direct use if needed
func GetLogger() *zap.Logger {
	if log == nil {
		// Initialize default logger if not initialized
		_ = InitLogger("info")
	}
	return log
}

// Info logs an info message
func Info(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		log.Info(msg, zap.Any("fields", fields[0]))
	} else {
		log.Info(msg)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		log.Debug(msg, zap.Any("fields", fields[0]))
	} else {
		log.Debug(msg)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		log.Warn(msg, zap.Any("fields", fields[0]))
	} else {
		log.Warn(msg)
	}
}

// Error logs an error message
// Can be called as: Error(msg), Error(msg, err), or Error(msg, err, fields)
func Error(msg string, args ...interface{}) {
	if log == nil {
		return
	}

	var err error
	var fields interface{}

	// Parse arguments
	if len(args) > 0 {
		// First arg could be error or fields map
		if errVal, ok := args[0].(error); ok {
			err = errVal
			if len(args) > 1 {
				fields = args[1]
			}
		} else if mapVal, ok := args[0].(map[string]interface{}); ok {
			fields = mapVal
		}
	}

	// Build log call
	if err != nil {
		if fields != nil {
			log.Error(msg, zap.Error(err), zap.Any("fields", fields))
		} else {
			log.Error(msg, zap.Error(err))
		}
	} else {
		if fields != nil {
			log.Error(msg, zap.Any("fields", fields))
		} else {
			log.Error(msg)
		}
	}
}

// Fatal logs a fatal message and exits
// Can be called as: Fatal(msg), Fatal(msg, err), or Fatal(msg, err, fields)
func Fatal(msg string, args ...interface{}) {
	if log == nil {
		return
	}

	var err error
	var fields interface{}

	// Parse arguments
	if len(args) > 0 {
		// First arg could be error or fields map
		if errVal, ok := args[0].(error); ok {
			err = errVal
			if len(args) > 1 {
				fields = args[1]
			}
		} else if mapVal, ok := args[0].(map[string]interface{}); ok {
			fields = mapVal
		}
	}

	// Build log call
	if err != nil {
		if fields != nil {
			log.Fatal(msg, zap.Error(err), zap.Any("fields", fields))
		} else {
			log.Fatal(msg, zap.Error(err))
		}
	} else {
		if fields != nil {
			log.Fatal(msg, zap.Any("fields", fields))
		} else {
			log.Fatal(msg)
		}
	}
}

// WithField creates a map with a single key-value pair for structured logging
func WithField(key string, value interface{}) map[string]interface{} {
	return map[string]interface{}{key: value}
}

// WithFields creates a map from the given key-value pairs for structured logging
func WithFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}

// isTerminalOutput checks if stdout is connected to a terminal
func isTerminalOutput() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
