package config

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Port     string
	MongoURI string
	DBName   string
	MongoDB  *mongo.Client
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Port = os.Getenv("PORT")
	MongoURI = os.Getenv("MONGO_URI")
	DBName = os.Getenv("DB_NAME")

	clientOptions := options.Client().ApplyURI(MongoURI)

	MongoDB, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = MongoDB.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
}
