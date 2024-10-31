package main

import (
	"fmt"
	"github.com/fossyy/filekeeper/encryption"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/storage"
	"strconv"

	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/cache"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/routes/admin"
	"github.com/fossyy/filekeeper/routes/client"
	"github.com/fossyy/filekeeper/utils"
)

func main() {
	clientAddr := fmt.Sprintf("%s:%s", utils.Getenv("SERVER_HOST"), utils.Getenv("SERVER_PORT"))
	adminAddr := fmt.Sprintf("%s:%s", utils.Getenv("SERVER_HOST"), "27000")

	dbUser := utils.Getenv("DB_USERNAME")
	dbPass := utils.Getenv("DB_PASSWORD")
	dbHost := utils.Getenv("DB_HOST")
	dbPort := utils.Getenv("DB_PORT")
	dbName := utils.Getenv("DB_NAME")
	redisHost := utils.Getenv("REDIS_HOST")
	redisPort := utils.Getenv("REDIS_PORT")
	redisPassword := utils.Getenv("REDIS_PASSWORD")

	database := db.NewPostgresDB(dbUser, dbPass, dbHost, dbPort, dbName, db.DisableSSL)
	cacheServer := cache.NewRedisServer(redisHost, redisPort, redisPassword, database)

	smtpPort, _ := strconv.Atoi(utils.Getenv("SMTP_PORT"))
	mailServer := email.NewSmtpServer(utils.Getenv("SMTP_HOST"), smtpPort, utils.Getenv("SMTP_USER"), utils.Getenv("SMTP_PASSWORD"))

	bucket := utils.Getenv("S3_BUCKET_NAME")
	region := utils.Getenv("S3_REGION")
	endpoint := utils.Getenv("S3_ENDPOINT")
	accessKey := utils.Getenv("S3_ACCESS_KEY")
	secretKey := utils.Getenv("S3_SECRET_KEY")
	S3 := storage.NewS3(bucket, region, endpoint, accessKey, secretKey)

	app.Server = app.NewClientServer(clientAddr, middleware.Handler(client.SetupRoutes()), *logger.Logger(), database, cacheServer, encryption.NewAesEncryption(), S3, mailServer)
	app.Admin = app.NewAdminServer(adminAddr, middleware.Handler(admin.SetupRoutes()), database)

	go func() {
		fmt.Printf("Admin Web App Listening on http://%s\n", app.Admin.Addr)
		err := app.Admin.ListenAndServe()
		if err != nil {
			panic(err)
			return
		}
	}()

	fmt.Printf("Client Web App Listening on http://%s\n", app.Server.Addr)
	err := app.Server.ListenAndServe()
	if err != nil {
		panic(err)
		return
	}
}
