package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// LogLevel represents the severity of the log message
type LogLevel string

const (
	// Log levels
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     LogLevel    `json:"level"`
	Message   string      `json:"message"`
	Function  string      `json:"function,omitempty"`
	File      string      `json:"file,omitempty"`
	Line      int         `json:"line,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// log creates and outputs a log entry
func log(level LogLevel, message string, data interface{}) {
	// Get caller information
	pc, file, line, ok := runtime.Caller(2)
	funcName := "unknown"
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName = fn.Name()
		}
	}

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Function:  funcName,
		File:      file,
		Line:      line,
		Data:      data,
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling log entry: %v\n", err)
		return
	}

	// Print to stdout (Lambda will capture this for CloudWatch)
	fmt.Println(string(jsonBytes))

	// If fatal, exit the program
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func Debug(message string, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	log(DEBUG, message, logData)
}

// Info logs an info message
func Info(message string, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	log(INFO, message, logData)
}

// Warn logs a warning message
func Warn(message string, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	log(WARN, message, logData)
}

// Error logs an error message
func Error(message string, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	log(ERROR, message, logData)
}

// Fatal logs a fatal message and exits the program
func Fatal(message string, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	log(FATAL, message, logData)
}