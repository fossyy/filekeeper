package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type AggregatedLogger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	panicLogger *log.Logger
}

func init() {
	if _, err := os.Stat("log"); os.IsNotExist(err) {
		os.Mkdir("log", os.ModePerm)
	}
}

func Logger() *AggregatedLogger {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02-15-04")
	file, err := os.OpenFile(fmt.Sprintf("log/log-%s.log", formattedTime), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &AggregatedLogger{}
	}
	flag := log.Ldate | log.Ltime
	writer := io.MultiWriter(os.Stdout, file)
	infoLogger := log.New(writer, "INFO: ", flag)
	warnLogger := log.New(writer, "WARN: ", flag)
	errorLogger := log.New(writer, "ERROR: ", flag)
	panicLogger := log.New(writer, "PANIC: ", flag)

	return &AggregatedLogger{
		infoLogger:  infoLogger,
		warnLogger:  warnLogger,
		errorLogger: errorLogger,
		panicLogger: panicLogger,
	}
}

func (l *AggregatedLogger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

func (l *AggregatedLogger) Warn(v ...interface{}) {
	l.warnLogger.Println(v...)
}

func (l *AggregatedLogger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

func (l *AggregatedLogger) Panic(v ...interface{}) {
	l.panicLogger.Panic(v...)
}
