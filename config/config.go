package config

import (
	"context"
	"log"

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
	// Manually set the configuration values here
	Port = "80"
	MongoURI = "mongodb+srv://hiskiaparhusip:iaTXSU4ciZMCVz5E@cluster.ryttvxb.mongodb.net/?retryWrites=true&w=majority&appName=Cluster"
	DBName = "db_dilithium"

	clientOptions := options.Client().ApplyURI(MongoURI)

	var err error
	MongoDB, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = MongoDB.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
}
