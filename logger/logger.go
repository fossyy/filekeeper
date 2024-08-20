package logger

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var logFlag *bool
var infoLoggerWriter io.Writer

type AggregatedLogger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	panicLogger *log.Logger
}

func init() {
	logFlag = flag.Bool("log", false, "Enable logging")
	flag.Parse()

	if _, err := os.Stat("log"); os.IsNotExist(err) {
		os.Mkdir("log", os.ModePerm)
	}
}

func Logger() *AggregatedLogger {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02-15-04")
	file, err := os.OpenFile(fmt.Sprintf("log/log-%s.log", formattedTime), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	multiLogger := io.MultiWriter(os.Stdout, file)

	if err != nil {
		return &AggregatedLogger{}
	}

	if *logFlag {
		infoLoggerWriter = multiLogger
	} else {
		infoLoggerWriter = file
	}

	slug := log.Ldate | log.Ltime
	infoLogger := log.New(infoLoggerWriter, "INFO: ", slug)
	warnLogger := log.New(multiLogger, "WARN: ", slug)
	errorLogger := log.New(multiLogger, "ERROR: ", slug)
	panicLogger := log.New(multiLogger, "PANIC: ", slug)

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
