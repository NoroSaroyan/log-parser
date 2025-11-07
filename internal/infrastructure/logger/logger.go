package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// InitLogger initializes logrus with the specified level and CLI-friendly formatting
func InitLogger(level string) error {
	logger = logrus.New()
	
	// Set log level
	switch strings.ToUpper(level) {
	case "DEBUG":
		logger.SetLevel(logrus.DebugLevel)
	case "INFO":
		logger.SetLevel(logrus.InfoLevel)
	case "WARN":
		logger.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logger.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set output
	logger.SetOutput(os.Stdout)

	// Use colored text formatter for CLI
	if isTerminal() {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
		})
	}

	return nil
}

// SetOutput sets the logger output (useful for testing)
func SetOutput(output io.Writer) {
	if logger != nil {
		logger.SetOutput(output)
	}
}

// GetLogger returns the logger instance for direct use if needed
func GetLogger() *logrus.Logger {
	return logger
}

// Info logs an info message with optional structured fields
func Info(msg string, fields ...logrus.Fields) {
	if logger == nil {
		return
	}
	
	if len(fields) > 0 {
		logger.WithFields(fields[0]).Info(msg)
	} else {
		logger.Info(msg)
	}
}

// Debug logs a debug message with optional structured fields
func Debug(msg string, fields ...logrus.Fields) {
	if logger == nil {
		return
	}
	
	if len(fields) > 0 {
		logger.WithFields(fields[0]).Debug(msg)
	} else {
		logger.Debug(msg)
	}
}

// Warn logs a warning message with optional structured fields
func Warn(msg string, fields ...logrus.Fields) {
	if logger == nil {
		return
	}
	
	if len(fields) > 0 {
		logger.WithFields(fields[0]).Warn(msg)
	} else {
		logger.Warn(msg)
	}
}

// Error logs an error message with optional structured fields
func Error(msg string, fields ...logrus.Fields) {
	if logger == nil {
		return
	}
	
	if len(fields) > 0 {
		logger.WithFields(fields[0]).Error(msg)
	} else {
		logger.Error(msg)
	}
}

// Fatal logs a fatal message and exits the program
func Fatal(msg string, fields ...logrus.Fields) {
	if logger == nil {
		return
	}
	
	if len(fields) > 0 {
		logger.WithFields(fields[0]).Fatal(msg)
	} else {
		logger.Fatal(msg)
	}
}

// Helper functions to create logrus.Fields more easily
func Fields() logrus.Fields {
	return make(logrus.Fields)
}

func WithField(key string, value interface{}) logrus.Fields {
	return logrus.Fields{key: value}
}

func WithFields(fields map[string]interface{}) logrus.Fields {
	return logrus.Fields(fields)
}

// isTerminal checks if stdout is connected to a terminal (for color support)
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
