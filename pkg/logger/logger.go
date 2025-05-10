package logger

import (
	"log"
	"os"
)

// Logger represents a simple logger
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	isDebug     bool
}

// NewLogger creates a new logger instance
func NewLogger(debug bool) *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		isDebug:     debug,
	}
}

// Info logs info level message
func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

// Infof logs formatted info level message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error logs error level message
func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

// Errorf logs formatted error level message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug logs debug level message if debug mode is enabled
func (l *Logger) Debug(v ...interface{}) {
	if l.isDebug {
		l.debugLogger.Println(v...)
	}
}

// Debugf logs formatted debug level message if debug mode is enabled
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.isDebug {
		l.debugLogger.Printf(format, v...)
	}
}

// Fatal logs error message and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.errorLogger.Fatal(v...)
}

// Fatalf logs formatted error message and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errorLogger.Fatalf(format, v...)
}