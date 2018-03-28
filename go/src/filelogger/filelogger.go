package filelogger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Level of a log statement
type Level string

const (
	// DEBUG - extra debug information
	DEBUG Level = "Debug  "
	// INFO - informational
	INFO Level = "Info   "
	// WARNING - indicator of something going wrong
	WARNING Level = "Warning"
	// ERROR - something has gone wrong, but the application can continue
	ERROR Level = "Error  "
	// FATAL - something has gone wrong, and the application cannot continue
	FATAL Level = "Fatal  "
)

// State of the logger
type State int

const (
	// NORMAL - Print Info and above to console and disk
	NORMAL State = 0
	// QUIET - Print Nothing
	QUIET State = 1
	// NOWRITE - Normal Do not write to disk
	NOWRITE State = 2
	// ALL - Print all
	ALL State = 3
)

// Logger is a logger which can log to disk
type Logger struct {
	name  string
	log   *log.Logger
	file  *os.File
	state State
}

// NewFileLogger creates a new logger that may log to disk
func NewFileLogger(loggerName string, state State) (logger *Logger, err error) {
	// make logs folder if not existing already
	err = os.MkdirAll("logs", 0700)
	if err != nil {
		return nil, fmt.Errorf("unable to create log folder: %s", err)
	}
	// open file for writing
	f, err := os.Create("logs/" + loggerName + timeNow() + ".log")
	if err != nil {
		return nil, fmt.Errorf("unable to create log file: %s", err)
	}
	logger = &Logger{
		name:  loggerName,
		log:   log.New(os.Stderr, fmt.Sprintf("[%s] ", loggerName), log.Ltime|log.Lmicroseconds),
		file:  f,
		state: state,
	}
	return logger, nil
}

// Exit the logger
func (l *Logger) Exit() {
	l.file.Close()
}

// Log takes a level and some data to be logged per the logger state
func (l *Logger) Log(level Level, data string) {
	if l.file == nil || l.log == nil {
		fmt.Println("ERROR: Log is incorrectly initialized")
		return
	}

	logString := fmt.Sprintf("| %s | %s", level, data)
	switch l.state {
	case NOWRITE:
		// Do not write anything
	default:
		lineHeader := fmt.Sprintf("[ %s | %s ]", l.name, timeNow())
		_, err := l.file.WriteString(lineHeader + logString + "\n")
		if err != nil {
			fmt.Printf("write failed: %s\n", err)
		}
	}

	switch level {
	case DEBUG:
		if l.state == ALL {
			l.log.Print(logString)
		}
	default:
		if l.state != QUIET {
			l.log.Print(logString)
		}
	}
}

// Debug Level log
func (l *Logger) Debug(data string) {
	l.Log(DEBUG, data)
}

// Info Level log
func (l *Logger) Info(data string) {
	l.Log(INFO, data)
}

// Warning Level log
func (l *Logger) Warning(data string) {
	l.Log(WARNING, data)
}

// Error Level log
func (l *Logger) Error(data string) {
	l.Log(ERROR, data)
}

// Fatal Level log
func (l *Logger) Fatal(data string) {
	l.Log(FATAL, data)
}

func timeNow() string {
	return time.Now().Format("2006-01-02_15:04:05")
}
