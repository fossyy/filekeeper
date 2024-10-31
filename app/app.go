package app

import (
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"net/http"
)

var Server App
var Admin App

type App struct {
	http.Server
	Database   types.Database
	Cache      types.CachingServer
	Storage    types.Storage
	Encryption types.Encryption
	Logger     *logger.AggregatedLogger
	Mail       *email.SmtpServer
}

func NewClientServer(addr string, handler http.Handler, logger logger.AggregatedLogger, database types.Database, cache types.CachingServer, encryption types.Encryption, storage types.Storage, mail email.SmtpServer) App {
	return App{
		Server: http.Server{
			Addr:    addr,
			Handler: handler,
		},
		Storage:    storage,
		Logger:     &logger,
		Database:   database,
		Encryption: encryption,
		Cache:      cache,
		Mail:       &mail,
	}
}

func NewAdminServer(addr string, handler http.Handler, database types.Database) App {
	return App{
		Server: http.Server{
			Addr:    addr,
			Handler: handler,
		},
		// TODO: Remove the dummy struct
		Logger:   &logger.AggregatedLogger{},
		Database: database,
		Mail:     &email.SmtpServer{},
	}
}
