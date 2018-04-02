package logger

import (
	"filelogger/level"
	"filelogger/state"
	"fmt"
	"log"
	"os"
	"time"
)

/*
	It would be brilliant to support a format call like in fmt.Printf
	Currently most calls look like: logger.Info(fmt.Sprintf("this is an example %v", valuedthing))
*/

// Logger is a logger which can log to disk
type Logger struct {
	name  string
	log   *log.Logger
	file  *os.File
	state state.State
}

var globalLoggers = make(map[string]*Logger)

// NewFileLogger creates a new logger that may log to disk
func NewFileLogger(loggerName string, state state.State) (logger *Logger, err error) {
	if globalLoggers[loggerName] != nil {
		return globalLoggers[loggerName], nil
	}
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
	globalLoggers[loggerName] = logger
	return logger, nil
}

// GetLogger by loggerName if it exists, or create a new normal logger by that name
func GetLogger(loggerName string) (logger *Logger) {
	if globalLoggers[loggerName] != nil {
		return globalLoggers[loggerName]
	}
	logger, err := NewFileLogger(loggerName, state.NORMAL)
	if err != nil {
		fmt.Printf("logger error: unable to create new logger: %s", err)
		return nil
	}
	return logger
}

// Exit the logger
func (l *Logger) Exit() {
	l.file.Close()
}

// Log takes a level and some data to be logged per the logger state
func (l *Logger) Log(givenLevel level.Level, data string) {
	if l.file == nil || l.log == nil {
		fmt.Println("ERROR: Log is incorrectly initialized")
		return
	}

	logString := fmt.Sprintf("| %s | %s", givenLevel, data)
	switch l.state {
	case state.NOWRITE:
		// Do not write anything
	default:
		lineHeader := fmt.Sprintf("[ %s | %s ]", l.name, timeNow())
		_, err := l.file.WriteString(lineHeader + logString + "\n")
		if err != nil {
			fmt.Printf("write failed: %s\n", err)
		}
	}

	switch givenLevel {
	case level.DEBUG:
		if l.state == state.DEBUGGING {
			l.log.Print(logString)
		}
	case level.INFO:
		if l.state != state.QUIET {
			fmt.Println(data)
		}
	default:
		if l.state != state.QUIET {
			l.log.Print(logString)
		}
	}
}

// Debug Level log
func (l *Logger) Debug(data string) {
	l.Log(level.DEBUG, data)
}

// Info Level log
func (l *Logger) Info(data string) {
	l.Log(level.INFO, data)
}

// Warning Level log
func (l *Logger) Warning(data string) {
	l.Log(level.WARNING, data)
}

// Error Level log
func (l *Logger) Error(data string) {
	l.Log(level.ERROR, data)
}

// Fatal Level log
func (l *Logger) Fatal(data string) {
	l.Log(level.FATAL, data)
}

func timeNow() string {
	return time.Now().Format("2006-01-02_15:04:05")
}
