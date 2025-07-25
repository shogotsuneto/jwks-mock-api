package logger

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger wraps the standard log package with level support
type Logger struct {
	level LogLevel
}

var defaultLogger *Logger

// Init initializes the default logger with the specified level
func Init(levelStr string) {
	level := parseLogLevel(levelStr)
	defaultLogger = &Logger{level: level}
}

// parseLogLevel converts a string to LogLevel
func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO // default to info if invalid level
	}
}

// shouldLog determines if a message should be logged based on the current level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(DEBUG) {
		log.Print(append([]interface{}{"[DEBUG] "}, v...)...)
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(DEBUG) {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message
func Info(v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(INFO) {
		log.Print(append([]interface{}{"[INFO] "}, v...)...)
	}
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(INFO) {
		log.Printf("[INFO] "+format, v...)
	}
}

// Printf is an alias for Infof to maintain compatibility with existing log.Printf calls
func Printf(format string, v ...interface{}) {
	Infof(format, v...)
}

// Print is an alias for Info to maintain compatibility with existing log.Print calls
func Print(v ...interface{}) {
	Info(v...)
}

// Println is an alias for Info to maintain compatibility with existing log.Println calls
func Println(v ...interface{}) {
	Info(v...)
}

// Warn logs a warning message
func Warn(v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(WARN) {
		log.Print(append([]interface{}{"[WARN] "}, v...)...)
	}
}

// Warnf logs a formatted warning message
func Warnf(format string, v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(WARN) {
		log.Printf("[WARN] "+format, v...)
	}
}

// Error logs an error message
func Error(v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(ERROR) {
		log.Print(append([]interface{}{"[ERROR] "}, v...)...)
	}
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	if defaultLogger != nil && defaultLogger.shouldLog(ERROR) {
		log.Printf("[ERROR] "+format, v...)
	}
}

// Fatal logs a fatal message and exits (always shown regardless of level)
func Fatal(v ...interface{}) {
	log.Fatal(append([]interface{}{"[FATAL] "}, v...)...)
}

// Fatalf logs a formatted fatal message and exits (always shown regardless of level)
func Fatalf(format string, v ...interface{}) {
	log.Fatalf("[FATAL] "+format, v...)
}

// SetOutput sets the output destination for the logger
func SetOutput(file *os.File) {
	log.SetOutput(file)
}
