package db

import (
	"fmt"
	"os"
	"strings"

	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

var log *logger.AggregatedLogger

func init() {
	var err error
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", utils.Getenv("DB_USERNAME"), utils.Getenv("DB_PASSWORD"), utils.Getenv("DB_HOST"), utils.Getenv("DB_PORT"), utils.Getenv("DB_NAME"))
	DB, err = gorm.Open(mysql.Open(connection), &gorm.Config{}, &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}
	file, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Error("Error opening file: %s", err.Error())
	}
	querys := strings.Split(string(file), "\n")
	for _, query := range querys {
		err := DB.Exec(query).Error
		if err != nil {
			panic(err.Error())
		}
	}
}
