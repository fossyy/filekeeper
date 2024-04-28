package app

import (
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/routes"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
)

type App struct {
	http.Server
	DB db.Database
}

var Server App

func NewServer(addr string, handler http.Handler, database db.Database) App {
	return App{
		Server: http.Server{
			Addr:    addr,
			Handler: handler,
		},
		DB: database,
	}
}

func Start() {
	serverAddr := fmt.Sprintf("%s:%s", utils.Getenv("SERVER_HOST"), utils.Getenv("SERVER_PORT"))

	dbUser := utils.Getenv("DB_USERNAME")
	dbPass := utils.Getenv("DB_PASSWORD")
	dbHost := utils.Getenv("DB_HOST")
	dbPort := utils.Getenv("DB_PORT")
	dbName := utils.Getenv("DB_NAME")

	database := db.NewMYSQLdb(dbUser, dbPass, dbHost, dbPort, dbName)
	db.DB = database

	Server = NewServer(serverAddr, middleware.Handler(routes.SetupRoutes()), database)
	fmt.Printf("Listening on http://%s\n", Server.Addr)
	err := Server.ListenAndServe()
	if err != nil {
		panic(err)
		return
	}
}
