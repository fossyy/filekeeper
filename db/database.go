package db

import (
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"os"
	"strings"
)

var log *logger.AggregatedLogger
var DB *gorm.DB

type mySQLdb struct {
	*gorm.DB
}

type Database interface {
	IsUserRegistered(email string, username string) bool

	CreateUser(user *models.User) error
	GetUser(email string) (*models.User, error)
	UpdateUserPassword(email string, password string) error

	CreateFile(file *models.File) error
	GetFile(fileID string) (*models.File, error)
	GetUserFile(name string, ownerID string) (*models.File, error)
	GetFiles(ownerID string) ([]*models.File, error)

	CreateUploadInfo(info models.FilesUploaded) error
	GetUploadInfo(uploadID string) (*models.FilesUploaded, error)
	UpdateUpdateIndex(index int, fileID string)
	FinalizeFileUpload(fileID string)
}

func NewMYSQLdb(username, password, host, port, dbName string) Database {
	var err error
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
	DB, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       connection,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	file, err := os.ReadFile("schema.sql")
	if err != nil {
		panic("Error opening file: " + err.Error())
	}

	queries := strings.Split(string(file), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		err := DB.Exec(query).Error
		if err != nil {
			panic("Error executing query: " + err.Error())
		}
	}

	return &mySQLdb{DB}
}

func (db *mySQLdb) IsUserRegistered(email string, username string) bool {
	var data models.User
	err := db.DB.Table("users").Where("email = ? OR username = ?", email, username).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		return true
	}
	return true
}

func (db *mySQLdb) CreateUser(user *models.User) error {
	err := db.DB.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *mySQLdb) GetUser(email string) (*models.User, error) {
	var user models.User
	err := db.DB.Table("users").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *mySQLdb) UpdateUserPassword(email string, password string) error {
	err := db.DB.Table("users").Where("email = ?", email).Update("password", password).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *mySQLdb) CreateFile(file *models.File) error {
	err := db.DB.Create(file).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *mySQLdb) GetFile(fileID string) (*models.File, error) {
	var file models.File
	err := db.DB.Table("files").Where("id = ?", fileID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (db *mySQLdb) GetUserFile(name string, ownerID string) (*models.File, error) {
	var file models.File
	err := db.DB.Table("files").Where("name = ? AND owner_id = ?", name, ownerID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (db *mySQLdb) GetFiles(ownerID string) ([]*models.File, error) {
	var files []*models.File
	err := db.DB.Table("files").Where("owner_id = ?", ownerID).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, err
}

// CreateUploadInfo It's not optimal, but it's okay for now. Consider implementing caching instead of pushing all updates to the database for better performance in the future.
func (db *mySQLdb) CreateUploadInfo(info models.FilesUploaded) error {
	err := db.DB.Create(info).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *mySQLdb) GetUploadInfo(fileID string) (*models.FilesUploaded, error) {
	var info models.FilesUploaded
	err := db.DB.Table("files_uploadeds").Where("file_id = ?", fileID).First(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (db *mySQLdb) UpdateUpdateIndex(index int, fileID string) {
	db.DB.Table("files_uploadeds").Where("file_id = ?", fileID).Updates(map[string]interface{}{
		"Uploaded": index,
	})
}

func (db *mySQLdb) FinalizeFileUpload(fileID string) {
	db.DB.Table("files_uploadeds").Where("file_id = ?", fileID).Updates(map[string]interface{}{
		"Done": true,
	})
}
