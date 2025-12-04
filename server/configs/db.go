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
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		log.Fatal(err)
	}
	// Giả sử tên DB là NoteApp
	DB = client.Database(os.Getenv("DB_NAME"))
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}
