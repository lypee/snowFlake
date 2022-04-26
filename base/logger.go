package base

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level int8

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
)

var levelMap = map[Level]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARNING: "WARNING",
	ERROR:   "ERROR",
}

type Logger interface {
	DebugF(string, ...interface{})
	InfoF(string, ...interface{})
	WarningF(string, ...interface{})
	ErrorF(string, ...interface{})

	log(Level, string, ...interface{})
}

func DefaultLogger() Logger {
	return new(defaultLogger)
}

var DLogger = DefaultLogger()

type defaultLogger struct{}

func (d *defaultLogger) log(level Level, format string, a ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	sections := strings.Split(file, "/")
	file = sections[len(sections)-1]
	s := fmt.Sprintf("%s %s:%d %s %s", time.Now().String()[:19], file, line, levelMap[level], fmt.Sprintf(format, a...))
	fmt.Println(s)
	if level > WARNING {
		os.Exit(-1)
	}
}

func (d *defaultLogger) DebugF(format string, a ...interface{}) {
	d.log(DEBUG, format, a...)
}

func (d *defaultLogger) InfoF(format string, a ...interface{}) {
	d.log(INFO, format, a...)
}

func (d *defaultLogger) WarningF(format string, a ...interface{}) {
	d.log(WARNING, format, a...)
}

func (d *defaultLogger) ErrorF(format string, a ...interface{}) {
	d.log(ERROR, format, a...)
}

func DebugF(format string, a ...interface{}) {
	if DLogger == nil {
		DLogger = DefaultLogger()
	}
	DLogger.DebugF(format, a)
}

func InfoF(format string, a ...interface{}) {
	if DLogger == nil {
		DLogger = DefaultLogger()
	}
	DLogger.InfoF(format, a)
}

func WarningF(format string, a ...interface{}) {
	if DLogger == nil {
		DLogger = DefaultLogger()
	}
	DLogger.WarningF(format, a)
}

func ErrorF(format string, a ...interface{}) {
	if DLogger == nil {
		DLogger = DefaultLogger()
	}
	DLogger.ErrorF(format, a)
}
