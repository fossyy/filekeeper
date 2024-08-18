package main

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/routes"
	"github.com/fossyy/filekeeper/utils"
	"strconv"
)

func main() {
	serverAddr := fmt.Sprintf("%s:%s", utils.Getenv("SERVER_HOST"), utils.Getenv("SERVER_PORT"))

	dbUser := utils.Getenv("DB_USERNAME")
	dbPass := utils.Getenv("DB_PASSWORD")
	dbHost := utils.Getenv("DB_HOST")
	dbPort := utils.Getenv("DB_PORT")
	dbName := utils.Getenv("DB_NAME")

	database := db.NewPostgresDB(dbUser, dbPass, dbHost, dbPort, dbName, db.DisableSSL)
	db.DB = database

	smtpPort, _ := strconv.Atoi(utils.Getenv("SMTP_PORT"))
	mailServer := email.NewSmtpServer(utils.Getenv("SMTP_HOST"), smtpPort, utils.Getenv("SMTP_USER"), utils.Getenv("SMTP_PASSWORD"))

	app.Server = app.NewServer(serverAddr, middleware.Handler(routes.SetupRoutes()), *logger.Logger(), database, mailServer)
	fmt.Printf("Listening on http://%s\n", app.Server.Addr)
	err := app.Server.ListenAndServe()
	if err != nil {
		panic(err)
		return
	}
}
