package configs

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Giúp các package khác gọi nó
var DB *mongo.Client

func ConnectDatabasse() *mongo.Client {

	// Tạo context chỉ để kết nối và đặt timeout cho quá trình kết nối database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// defer đảm bảo cancel được thực hiện cuối hàm
	defer cancel()

	// Lấy connection string của mongodb
	uri := os.Getenv("CONNECTION_STRING")

	// Tạo mongo.Client
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to create client: ", err)
	}

	// Ping kiểm tra kết nối
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("Failed to ping cluster: ", err)
	}

	log.Println("Connected to MongoDB successfully!")

	// trả về client
	return client
}

// Lấy nhanh một collection
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("gfg").Collection(collectionName)
}
