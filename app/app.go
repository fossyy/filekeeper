package app

import (
	"net/http"

	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
)

var Server App

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
