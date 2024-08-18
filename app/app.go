package app

import (
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"net/http"
)

var Server App

type App struct {
	http.Server
	DB     *db.Database
	Logger *logger.AggregatedLogger
	Mail   *email.SmtpServer
}

func NewServer(addr string, handler http.Handler, logger logger.AggregatedLogger, database db.Database, mail email.SmtpServer) App {
	return App{
		Server: http.Server{
			Addr:    addr,
			Handler: handler,
		},
		Logger: &logger,
		DB:     &database,
		Mail:   &mail,
	}
}
