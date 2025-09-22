package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity of log messages
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

var (
	errorLogger *log.Logger
	warnLogger  *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger
	logFile     *os.File
	currentLogLevel = LogLevelInfo // Default to Info level
)

// InitLogger initializes the logging system with file output
func InitLogger(logDir string, logLevel LogLevel) error {
	currentLogLevel = logLevel

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("bot-%s.log", timestamp)
	logPath := filepath.Join(logDir, logFileName)

	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create writers that output to both file and stdout for different levels
	errorWriter := io.MultiWriter(logFile, os.Stderr)
	warnWriter := io.MultiWriter(logFile, os.Stdout)
	infoWriter := logFile // Info and debug only to file
	debugWriter := logFile

	// Initialize loggers with appropriate prefixes
	errorLogger = log.New(errorWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(warnWriter, "[WARN]  ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(infoWriter, "[INFO]  ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(debugWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	LogInfo("Logger initialized - Level: %v, File: %s", logLevel, logPath)
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogError logs error messages (always visible)
func LogError(format string, args ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, args...)
	}
}

// LogWarn logs warning messages
func LogWarn(format string, args ...interface{}) {
	if currentLogLevel >= LogLevelWarn && warnLogger != nil {
		warnLogger.Printf(format, args...)
	}
}

// LogInfo logs info messages
func LogInfo(format string, args ...interface{}) {
	if currentLogLevel >= LogLevelInfo && infoLogger != nil {
		infoLogger.Printf(format, args...)
	}
}

// LogDebug logs debug messages
func LogDebug(format string, args ...interface{}) {
	if currentLogLevel >= LogLevelDebug && debugLogger != nil {
		debugLogger.Printf(format, args...)
	}
}

// GetLogLevelFromString converts string to LogLevel
func GetLogLevelFromString(level string) LogLevel {
	switch level {
	case "error":
		return LogLevelError
	case "warn":
		return LogLevelWarn
	case "info":
		return LogLevelInfo
	case "debug":
		return LogLevelDebug
	default:
		return LogLevelInfo
	}
}