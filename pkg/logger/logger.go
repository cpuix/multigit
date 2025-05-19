// Package logger provides a configurable logging interface for the multigit application.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// LevelDebug represents debug log level
	LevelDebug LogLevel = iota
	// LevelInfo represents info log level
	LevelInfo
	// LevelWarn represents warning log level
	LevelWarn
	// LevelError represents error log level
	LevelError
	// LevelFatal represents fatal log level
	LevelFatal
)

var (
	// DefaultLevel is the default log level
	DefaultLevel = LevelInfo

	// DefaultOutput is the default output for logs
	DefaultOutput = os.Stdout

	// DefaultFlags are the default log flags
	DefaultFlags = log.LstdFlags | log.Lshortfile

	// DefaultCallerSkip is the number of stack frames to skip when reporting the file and line number
	DefaultCallerSkip = 3

	// colors for different log levels
	debugColor = color.New(color.FgBlue)
	infoColor  = color.New(color.FgGreen)
	warnColor  = color.New(color.FgYellow)
errorColor = color.New(color.FgRed)
	fatalColor = color.New(color.BgRed, color.FgWhite)
)

// Logger represents a logger instance
type Logger struct {
	*log.Logger
	level      LogLevel
	mu         sync.Mutex
	callerSkip int
}

// New creates a new logger instance
func New(out io.Writer, prefix string, flag int, level LogLevel) *Logger {
	return &Logger{
		Logger:     log.New(out, prefix, flag),
		level:      level,
		callerSkip: DefaultCallerSkip,
	}
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetCallerSkip sets the number of stack frames to skip when reporting the file and line number
func (l *Logger) SetCallerSkip(skip int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.callerSkip = skip
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	if l.level <= LevelDebug {
		l.Output(2, debugColor.Sprint(append([]interface{}{"[DEBUG] "}, v...)...))
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= LevelDebug {
		l.Output(2, debugColor.Sprintf("[DEBUG] "+format, v...))
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	if l.level <= LevelInfo {
		l.Output(2, infoColor.Sprint(append([]interface{}{"[INFO]  "}, v...)...))
	}
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= LevelInfo {
		l.Output(2, infoColor.Sprintf("[INFO]  "+format, v...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(v ...interface{}) {
	if l.level <= LevelWarn {
		l.Output(2, warnColor.Sprint(append([]interface{}{"[WARN]  "}, v...)...))
	}
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= LevelWarn {
		l.Output(2, warnColor.Sprintf("[WARN]  "+format, v...))
	}
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	if l.level <= LevelError {
		l.Output(2, errorColor.Sprint(append([]interface{}{"[ERROR] "}, v...)...))
	}
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= LevelError {
		l.Output(2, errorColor.Sprintf("[ERROR] "+format, v...))
	}
}

// Fatal logs a fatal message and exits the application
func (l *Logger) Fatal(v ...interface{}) {
	l.Output(2, fatalColor.Sprint(append([]interface{}{"[FATAL] "}, v...)...))
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits the application
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(2, fatalColor.Sprintf("[FATAL] "+format, v...))
	os.Exit(1)
}

// WithCaller adds caller information to the log message
func (l *Logger) WithCaller() *Logger {
	_, file, line, ok := runtime.Caller(l.callerSkip)
	if !ok {
		file = "???"
		line = 0
	}

	// Get just the filename from the full path
	_, filename := filepath.Split(file)

	caller := fmt.Sprintf("%s:%d", filename, line)

	return &Logger{
		Logger:     log.New(l.Writer(), caller+" ", l.Flags()),
		level:      l.level,
		callerSkip: l.callerSkip,
	}
}

// ParseLevel parses a log level string into a LogLevel
func ParseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Default logger instance
var defaultLogger = New(DefaultOutput, "", DefaultFlags, DefaultLevel)

// SetDefaultLogger sets the default logger
func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

// SetLevel sets the log level for the default logger
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// Debug logs a debug message using the default logger
func Debug(v ...interface{}) {
	defaultLogger.Output(2, debugColor.Sprint(append([]interface{}{"[DEBUG] "}, v...)...))
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format string, v ...interface{}) {
	defaultLogger.Output(2, debugColor.Sprintf("[DEBUG] "+format, v...))
}

// Info logs an info message using the default logger
func Info(v ...interface{}) {
	defaultLogger.Output(2, infoColor.Sprint(append([]interface{}{"[INFO]  "}, v...)...))
}

// Infof logs a formatted info message using the default logger
func Infof(format string, v ...interface{}) {
	defaultLogger.Output(2, infoColor.Sprintf("[INFO]  "+format, v...))
}

// Warn logs a warning message using the default logger
func Warn(v ...interface{}) {
	defaultLogger.Output(2, warnColor.Sprint(append([]interface{}{"[WARN]  "}, v...)...))
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, v ...interface{}) {
	defaultLogger.Output(2, warnColor.Sprintf("[WARN]  "+format, v...))
}

// Error logs an error message using the default logger
func Error(v ...interface{}) {
	defaultLogger.Output(2, errorColor.Sprint(append([]interface{}{"[ERROR] "}, v...)...))
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, v ...interface{}) {
	defaultLogger.Output(2, errorColor.Sprintf("[ERROR] "+format, v...))
}

// Fatal logs a fatal message using the default logger and exits the application
func Fatal(v ...interface{}) {
	defaultLogger.Output(2, fatalColor.Sprint(append([]interface{}{"[FATAL] "}, v...)...))
	os.Exit(1)
}

// Fatalf logs a formatted fatal message using the default logger and exits the application
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Output(2, fatalColor.Sprintf("[FATAL] "+format, v...))
	os.Exit(1)
}
