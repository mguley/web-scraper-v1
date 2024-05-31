package utils

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
)

const (
	Filename = "app.log"
	LogDir   = "logs"
)

// Logger is the interface for the custom logger.
// It defines the methods for logging error and informational messages.
type Logger interface {
	// LogError logs an error message along with the stack trace.
	//
	// Parameters:
	// - err: The error to log.
	// - includeCallerInfo: If provided and true, includes the file and line number of the caller in the log.
	LogError(err error, includeCallerInfo ...bool)

	// LogInfo logs an informational message.
	//
	// Parameters:
	// - message: The informational message to log.
	// - includeCallerInfo: If provided and true, includes the file and line number of the caller in the log.
	LogInfo(message string, includeCallerInfo ...bool)

	// Close closes the logger, releasing any open resources.
	Close() error
}

// logger is a concrete implementation of the Logger interface.
// It embeds a standard library log.Logger to handle logging to a file.
type logger struct {
	log     *log.Logger
	logFile *os.File
}

// instance holds the singleton instance of the logger.
var (
	instance *logger
	once     sync.Once
)

// GetLogger returns the singleton instance of the logger.
// It initializes the logger on the first call and returns the same instance on subsequent calls.
// This ensures that there is only one logger instance throughout the application lifecycle.
func GetLogger() Logger {
	once.Do(func() {
		instance = &logger{}
		instance.init()
	})
	return instance
}

// init initializes the logger instance.
// It ensures the logs directory exists and opens the log file for appending logs.
// If the log file cannot be opened, it fails back to logging to standard output (stdout).
func (logger *logger) init() {
	// Ensure the logs directory exists
	if _, err := os.Stat(LogDir); os.IsNotExist(err) {
		if createErr := os.Mkdir(LogDir, os.ModePerm); createErr != nil {
			log.Fatalf("failed to create log directory: %v", createErr)
		}
	}

	// Open the log file
	logFile, err := os.OpenFile(filepath.Join(LogDir, Filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Fallback to standard output if log file cannot be opened
		log.Printf("failed to open log file, using standard output: %v", err)
		logger.log = log.New(os.Stdout, "APP_LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
		return
	}

	logger.logFile = logFile
	// Set the logger to write to the log file
	logger.log = log.New(logFile, "APP_LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogError logs an error message along with the stack trace.
// It captures the current stack trace to provide context for the error, making it easier to debug issues.
//
// Parameters:
// - err: The error to log.
// - includeCallerInfo: If provided and true, includes the file and line number of the caller in the log.
func (logger *logger) LogError(err error, includeCallerInfo ...bool) {
	if err != nil {
		stackTrace := string(debug.Stack())
		logger.logMessage("ERROR", err.Error(), stackTrace, includeCallerInfo...)
	}
}

// LogInfo logs an informational message.
// This can be used to log general application events, such as startup and shutdown messages,
// or any other significant actions that occur during the application's operation.
//
// Parameters:
// - message: The informational message to log.
// - includeCallerInfo: If provided and true, includes the file and line number of the caller in the log.
func (logger *logger) LogInfo(message string, includeCallerInfo ...bool) {
	logger.logMessage("INFO", message, "", includeCallerInfo...)
}

// logMessage handles the actual logging, including optional caller info and stack trace.
// It is used by both LogError and LogInfo to avoid code duplication.
func (logger *logger) logMessage(logType string, message string, stackTrace string, includeCallerInfo ...bool) {
	includeCaller := false
	if len(includeCallerInfo) > 0 {
		includeCaller = includeCallerInfo[0]
	}

	if includeCaller {
		file, line := getCallerInfo()
		logger.log.Printf("\n%s: %s\nFILE: %s, LINE: %d\nSTACK TRACE:\n%s",
			logType, message, file, line, stackTrace)
	} else {
		logger.log.Printf("\n%s: %s\n", logType, message)
	}
}

// Close closes the logger, releasing any open resources.
func (logger *logger) Close() error {
	if logger.logFile != nil {
		return logger.logFile.Close()
	}
	return nil
}

// getCallerInfo retrieves the file and line number of the caller.
func getCallerInfo() (string, int) {
	// Skip the first two callers to get the actual caller of the logging function
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Extract the file name from the full path
	file = filepath.Base(file)
	return file, line
}
