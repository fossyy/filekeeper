package logger

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

	if _, err := os.Stat("log"); os.IsNotExist(err) {
		os.Mkdir("log", os.ModePerm)
	}
}

func Logger() *AggregatedLogger {
	flag.Parse()
	if err := os.MkdirAll("log", 0755); err != nil {
		return fallbackLogger()
	}
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02-15-04")
	file, err := os.OpenFile(fmt.Sprintf("log/log-%s.log", formattedTime), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	multiLogger := io.MultiWriter(os.Stdout, file)

	if err != nil {
		return fallbackLogger()
	}

	if *logFlag {
		infoLoggerWriter = multiLogger
	} else {
		infoLoggerWriter = file
	}

	slug := log.Ldate | log.Ltime
	infoLogger := log.New(infoLoggerWriter, "INFO: ", slug)

	return &AggregatedLogger{
		infoLogger:  infoLogger,
		warnLogger:  log.New(multiLogger, "WARN: ", slug),
		errorLogger: log.New(multiLogger, "ERROR: ", slug),
		panicLogger: log.New(multiLogger, "PANIC: ", slug),
	}
}

func fallbackLogger() *AggregatedLogger {
	return &AggregatedLogger{
		infoLogger:  log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime),
		warnLogger:  log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
		panicLogger: log.New(os.Stderr, "PANIC: ", log.Ldate|log.Ltime),
	}
}

func getFileAndLine() string {
	_, file, line, _ := runtime.Caller(2)
	filename := filepath.Base(file)
	return fmt.Sprintf("%s:%d", filename, line)
}

func (l *AggregatedLogger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

func (l *AggregatedLogger) Warn(v ...interface{}) {
	l.warnLogger.Println(getFileAndLine(), fmt.Sprint(v...))
}

func (l *AggregatedLogger) Error(v ...interface{}) {
	l.errorLogger.Println(getFileAndLine(), fmt.Sprint(v...))
}

func (l *AggregatedLogger) Panic(v ...interface{}) {
	l.panicLogger.Panic(getFileAndLine(), fmt.Sprint(v...))
}
