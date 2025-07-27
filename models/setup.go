package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func SetupTestDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		fmt.Println("Cannot connect to database sqlite ")
		log.Fatal("connection error:", err)
	}

	DB.Begin()
	DB.AutoMigrate(
		&User{},
		&Major{},
		&OptionTable{},
		&QuestionTable{},
		&QuizTable{},
		&UserResponse{},
		&EmailDomains{},
		&EmailsVerification{},
		&UsersImages{},
		&Connection{},
		&ConnectionRequest{},
		&Message{},
	)
}

func TearDownTestDB() {
	_, err := DB.DB()
	if err != nil {
		log.Fatal("connection error:", err)
	}
	DB.Rollback()
}

func ConnectDataBase() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	Dbdriver := os.Getenv("DB_DRIVER")
	DbHost := os.Getenv("DB_HOST")
	DbUser := os.Getenv("DB_USER")
	DbPassword := os.Getenv("DB_PASSWORD")
	DbName := os.Getenv("DB_NAME")
	DbPort := os.Getenv("DB_PORT")

	DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

	var err error
	DB, err = gorm.Open(mysql.Open(DBURL), &gorm.Config{
		Logger: newLogger,
	})

	log.Println("Database Host:", DbHost)

	if err != nil {
		fmt.Println("Cannot connect to database ", Dbdriver)
		log.Fatal("connection error:", err)
	} else {
		fmt.Println("We are connected to the database ", Dbdriver)
	}

	// DB.AutoMigrate(
	// 	&User{},
	// 	&Major{},
	// 	&OptionTable{},
	// 	&QuestionTable{},
	// 	&QuizTable{},
	// 	&UserResponse{},
	// 	&EmailDomains{},
	// 	&EmailsVerification{},
	// 	&UsersImages{},
	// 	&Connection{},
	// 	&ConnectionRequest{},
	// 	&Message{},
	// )
}
