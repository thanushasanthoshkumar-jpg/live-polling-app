package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"live-polling-backend/config"
)

var (
	MongoClient *mongo.Client
	DB          *mongo.Database
)

func ConnectMongo(cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoURL)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}

	MongoClient = client
	DB = client.Database(cfg.MongoDBName)

	log.Println("connected to MongoDB")
	ensureIndexes(ctx)
}

func UsersCollection() *mongo.Collection {
	return DB.Collection("users")
}

func PollsCollection() *mongo.Collection {
	return DB.Collection("polls")
}

func VotesCollection() *mongo.Collection {
	return DB.Collection("votes")
}
