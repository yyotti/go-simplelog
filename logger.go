package slog

import (
	"io"
	"log"
	"os"
	"sync"
)

// Logger : Set of loggers for debug/info/error log
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

var (
	debugPrefix = "[debug] "
	infoPrefix  = "[info]  "
	errorPrefix = "[error] "
)

var (
	l  *Logger
	mu sync.Mutex
)

func init() {
	initDefault()
}

func initDefault() {
	// Initialize default logger
	l = &Logger{
		infoLogger:  log.New(os.Stdout, infoPrefix, log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, errorPrefix, log.Ldate|log.Ltime),
	}
}

// Init initializes a Logger instance
func Init(debug bool, flag int, files ...io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	infoWriters := append([]io.Writer{os.Stdout}, files...)
	errorWriters := append([]io.Writer{os.Stderr}, files...)

	l = &Logger{
		infoLogger:  log.New(io.MultiWriter(infoWriters...), infoPrefix, flag),
		errorLogger: log.New(io.MultiWriter(errorWriters...), errorPrefix, flag),
	}

	if debug {
		l.debugLogger = log.New(os.Stdout, debugPrefix, flag)
	}
}

// Debug calls log.Print() on debug mode.
func Debug(v ...interface{}) {
	write(l.debugLogger, v...)
}

// Debugf calls log.Printf() on debug mode.
func Debugf(format string, v ...interface{}) {
	writef(l.debugLogger, format, v...)
}

// Debugln calls log.Println() on debug mode.
func Debugln(v ...interface{}) {
	writeln(l.debugLogger, v...)
}

// Info calls log.Print().
func Info(v ...interface{}) {
	write(l.infoLogger, v...)
}

// Infof calls log.Printf().
func Infof(format string, v ...interface{}) {
	writef(l.infoLogger, format, v...)
}

// Infoln calls log.Println().
func Infoln(v ...interface{}) {
	writeln(l.infoLogger, v...)
}

// Error calls log.Print().
func Error(v ...interface{}) {
	write(l.errorLogger, v...)
}

// Errorf calls log.Printf().
func Errorf(format string, v ...interface{}) {
	writef(l.errorLogger, format, v...)
}

// Errorln calls log.Println().
func Errorln(v ...interface{}) {
	writeln(l.errorLogger, v...)
}

func write(logger *log.Logger, v ...interface{}) {
	if logger == nil {
		return
	}

	logger.Print(v...)
}

func writef(logger *log.Logger, format string, v ...interface{}) {
	if logger == nil {
		return
	}

	logger.Printf(format, v...)
}

func writeln(logger *log.Logger, v ...interface{}) {
	if logger == nil {
		return
	}

	logger.Println(v...)
}

// SetDebugPrefix calls log.SetPrefix().
func SetDebugPrefix(prefix string) {
	setPrefix(l.debugLogger, prefix)
}

// SetInfoPrefix calls log.SetPrefix().
func SetInfoPrefix(prefix string) {
	setPrefix(l.infoLogger, prefix)
}

// SetErrorPrefix calls log.SetPrefix().
func SetErrorPrefix(prefix string) {
	setPrefix(l.errorLogger, prefix)
}

func setPrefix(logger *log.Logger, prefix string) {
	if logger == nil {
		return
	}

	logger.SetPrefix(prefix)
}
