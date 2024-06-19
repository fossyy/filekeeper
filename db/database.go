package db

import (
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/types/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"os"
	"strings"
)

var DB Database

type mySQLdb struct {
	*gorm.DB
}

type postgresDB struct {
	*gorm.DB
}

type SSLMode string

const (
	DisableSSL SSLMode = "disable"
	EnableSSL  SSLMode = "enable"
)

type Database interface {
	IsUserRegistered(email string, username string) bool
	IsEmailRegistered(email string) bool

	CreateUser(user *models.User) error
	GetUser(email string) (*models.User, error)
	UpdateUserPassword(email string, password string) error

	CreateFile(file *models.File) error
	GetFile(fileID string) (*models.File, error)
	GetUserFile(name string, ownerID string) (*models.File, error)
	GetFiles(ownerID string) ([]*models.File, error)

	UpdateUploadedByte(index int64, fileID string)
	UpdateUploadedChunk(index int64, fileID string)
	FinalizeFileUpload(fileID string)

	InitializeTotp(email string, secret string) error
}

func NewMYSQLdb(username, password, host, port, dbName string) Database {
	var err error
	var count int64

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	initDB, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connection,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	initDB.Raw("SELECT count(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", dbName).Scan(&count)
	if count <= 0 {
		if err := initDB.Exec("CREATE DATABASE IF NOT EXISTS " + dbName).Error; err != nil {
			panic("Error creating database: " + err.Error())
		}
	}

	connection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
	DB, err := gorm.Open(mysql.New(mysql.Config{
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

func NewPostgresDB(username, password, host, port, dbName string, mode SSLMode) Database {
	var err error
	var count int64

	connection := fmt.Sprintf("host=%s user=%s password=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", host, username, password, port, mode)
	initDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN: connection,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	initDB.Raw("SELECT count(*) FROM pg_database WHERE datname = ?", dbName).Scan(&count)
	if count <= 0 {
		if err := initDB.Exec("CREATE DATABASE " + dbName).Error; err != nil {
			panic("Error creating database: " + err.Error())
		}
	}

	connection = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", host, username, password, dbName, port, mode)
	DB, err := gorm.Open(postgres.New(postgres.Config{
		DSN: connection,
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

	return &postgresDB{DB}
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

func (db *mySQLdb) IsEmailRegistered(email string) bool {
	var data models.User
	err := db.DB.Table("users").Where("email = ? ", email).First(&data).Error
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
	var user models.User
	err := db.DB.Table("users").Where("email = ?", email).First(&user).Error
	if err != nil {
		return err
	}
	user.Password = password
	db.Save(&user)
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

func (db *mySQLdb) UpdateUploadedByte(byte int64, fileID string) {
	var file models.File
	db.DB.Table("files").Where("id = ?", fileID).First(&file)
	file.UploadedByte = byte
	db.Save(&file)
}

func (db *mySQLdb) UpdateUploadedChunk(index int64, fileID string) {
	var file models.File
	db.DB.Table("files").Where("id = ?", fileID).First(&file)
	file.UploadedChunk = index
	db.Save(&file)
}

func (db *mySQLdb) FinalizeFileUpload(fileID string) {
	var file models.File
	db.DB.Table("files").Where("id = ?", fileID).First(&file)
	file.Done = true
	db.Save(&file)
}

func (db *mySQLdb) InitializeTotp(email string, secret string) error {
	var user models.User
	err := db.DB.Table("users").Where("email = ?", email).First(&user).Error
	if err != nil {
		return err
	}
	user.Totp = secret
	err = db.Save(&user).Error
	if err != nil {
		return err
	}
	return nil
}

// POSTGRES FUNCTION
func (db *postgresDB) IsUserRegistered(email string, username string) bool {
	var data models.User
	err := db.DB.Table("users").Where("email = $1 OR username = $2", email, username).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		return true
	}
	return true
}

func (db *postgresDB) IsEmailRegistered(email string) bool {
	var data models.User
	err := db.DB.Table("users").Where("email = $1 ", email).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		return true
	}
	return true
}

func (db *postgresDB) CreateUser(user *models.User) error {
	err := db.DB.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *postgresDB) GetUser(email string) (*models.User, error) {
	var user models.User
	err := db.DB.Table("users").Where("email = $1", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *postgresDB) UpdateUserPassword(email string, password string) error {
	var user models.User
	err := db.DB.Table("users").Where("email = $1", email).First(&user).Error
	if err != nil {
		return err
	}
	user.Password = password
	db.Save(&user)
	return nil
}

func (db *postgresDB) CreateFile(file *models.File) error {
	err := db.DB.Create(file).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *postgresDB) GetFile(fileID string) (*models.File, error) {
	var file models.File
	err := db.DB.Table("files").Where("id = $1", fileID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (db *postgresDB) GetUserFile(name string, ownerID string) (*models.File, error) {
	var file models.File
	err := db.DB.Table("files").Where("name = $1 AND owner_id = $2", name, ownerID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (db *postgresDB) GetFiles(ownerID string) ([]*models.File, error) {
	var files []*models.File
	err := db.DB.Table("files").Where("owner_id = $1", ownerID).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, err
}

func (db *postgresDB) UpdateUploadedByte(byte int64, fileID string) {
	var file models.File
	db.DB.Table("files").Where("id = $1", fileID).First(&file)
	file.UploadedByte = byte
	db.Save(&file)
}
func (db *postgresDB) UpdateUploadedChunk(index int64, fileID string) {
	var file models.File
	db.DB.Table("files").Where("id = $1", fileID).First(&file)
	file.UploadedChunk = index
	db.Save(&file)
}

func (db *postgresDB) FinalizeFileUpload(fileID string) {
	var file models.File
	db.DB.Table("files").Where("id = $1", fileID).First(&file)
	file.Done = true
	db.Save(&file)
}

func (db *postgresDB) InitializeTotp(email string, secret string) error {
	var user models.User
	err := db.DB.Table("users").Where("email = $1", email).First(&user).Error
	if err != nil {
		return err
	}
	user.Totp = secret
	err = db.Save(&user).Error
	if err != nil {
		return err
	}
	return nil
}
