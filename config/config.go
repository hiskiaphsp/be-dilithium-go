package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	Port   string
	DBUser string
	DBPass string
	DBName string
	DBHost string
	DBPort string
	DB     *gorm.DB
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Port = os.Getenv("PORT")
	DBUser = os.Getenv("DB_USER")
	DBPass = os.Getenv("DB_PASS")
	DBName = os.Getenv("DB_NAME")
	DBHost = os.Getenv("DB_HOST")
	DBPort = os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DBUser, DBPass, DBHost, DBPort, DBName)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
}
