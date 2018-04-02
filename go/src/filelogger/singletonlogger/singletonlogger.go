package singletonlogger

import (
	"filelogger/level"
	"filelogger/logger"
	"filelogger/state"
	"fmt"
)

var singletonLogger *logger.Logger

// NewSingletonLogger creates a new global single instance of a logger.
func NewSingletonLogger(loggerName string, state state.State) (err error) {
	if singletonLogger != nil {
		return fmt.Errorf("logger already exists")
	}
	singletonLogger, err = logger.NewFileLogger(loggerName, state)
	if err != nil {
		return fmt.Errorf("unable to create a singletonlogger: %s", err)
	}
	return nil
}

// Debug Level log
func Debug(data string) {
	if singletonLogger == nil {
		fmt.Println("LOGGING ERROR: Singleton logger uninitialised!")
		fmt.Println(data)
		return
	}
	singletonLogger.Log(level.DEBUG, data)
}

// Info Level log
func Info(data string) {
	if singletonLogger == nil {
		fmt.Println("LOGGING ERROR: Singleton logger uninitialised!")
		fmt.Println(data)
		return
	}
	singletonLogger.Log(level.INFO, data)
}

// Warning Level log
func Warning(data string) {
	if singletonLogger == nil {
		fmt.Println("LOGGING ERROR: Singleton logger uninitialised!")
		fmt.Println(data)
		return
	}
	singletonLogger.Log(level.WARNING, data)
}

// Error Level log
func Error(data string) {
	if singletonLogger == nil {
		fmt.Println("LOGGING ERROR: Singleton logger uninitialised!")
		fmt.Println(data)
		return
	}
	singletonLogger.Log(level.ERROR, data)
}

// Fatal Level log
func Fatal(data string) {
	if singletonLogger == nil {
		fmt.Println("LOGGING ERROR: Singleton logger uninitialised!")
		fmt.Println(data)
		return
	}
	singletonLogger.Log(level.FATAL, data)
}
