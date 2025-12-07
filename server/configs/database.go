package configs

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("CONNECTION_STRING")))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatal("Error pinging MongoDB: ", err)
	}
	log.Println("Connected to MongoDB successfully")

	// Giả sử tên DB là NoteApp
	DB = client.Database(os.Getenv("DB_NAME"))
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}
