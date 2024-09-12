package db

import (
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

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

func NewMYSQLdb(username, password, host, port, dbName string) types.Database {
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

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

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

	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		panic(err.Error())
		return nil
	}
	err = DB.AutoMigrate(&models.File{})
	if err != nil {
		panic(err.Error())
		return nil
	}
	err = DB.AutoMigrate(&models.Allowance{})
	if err != nil {
		panic(err.Error())
		return nil
	}
	return &mySQLdb{DB}
}

func NewPostgresDB(username, password, host, port, dbName string, mode SSLMode) types.Database {
	var err error
	var count int64

	connection := fmt.Sprintf("host=%s user=%s password=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", host, username, password, port, mode)
	initDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN: connection,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

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

	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		panic(err.Error())
		return nil
	}
	err = DB.AutoMigrate(&models.File{})
	if err != nil {
		panic(err.Error())
		return nil
	}
	err = DB.AutoMigrate(&models.Allowance{})
	if err != nil {
		panic(err.Error())
		return nil
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
	err = db.CreateAllowance(user.UserID)
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

func (db *mySQLdb) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := db.DB.Table("users").Select("user_id, Username, Email").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
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

func (db *mySQLdb) CreateAllowance(userID uuid.UUID) error {
	userAllowance := &models.Allowance{
		UserID:        userID,
		AllowanceByte: 1024 * 1024 * 1024 * 10,
		AllowanceFile: 10,
	}
	err := db.DB.Create(userAllowance).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *mySQLdb) GetAllowance(userID uuid.UUID) (*models.Allowance, error) {
	var allowance models.Allowance
	err := db.DB.Table("allowances").Where("user_id = ?", userID).First(&allowance).Error
	if err != nil {
		return nil, err
	}
	return &allowance, nil
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
	err = db.CreateAllowance(user.UserID)
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

func (db *postgresDB) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := db.DB.Table("users").Select("user_id, username, email").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
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

func (db *postgresDB) CreateAllowance(userID uuid.UUID) error {
	userAllowance := &models.Allowance{
		UserID:        userID,
		AllowanceByte: 1024 * 1024 * 1024 * 10,
		AllowanceFile: 10,
	}
	err := db.DB.Create(userAllowance).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *postgresDB) GetAllowance(userID uuid.UUID) (*models.Allowance, error) {
	var allowance models.Allowance
	err := db.DB.Table("allowances").Where("user_id = $1", userID).First(&allowance).Error
	if err != nil {
		return nil, err
	}
	return &allowance, nil
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
