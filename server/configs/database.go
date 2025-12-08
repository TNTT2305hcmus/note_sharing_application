package configs

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() {
	// Lấy biến môi trường trong .env
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	if mongoURI == "" {
		log.Fatal("Chưa cấu hình MONGO_URI trong file .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kết nối đến db
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Lỗi khởi tạo client Mongo:", err)
	}

	// ping check
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Không thể ping tới MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB successfully")

	DB = client.Database(dbName)
}

func GetCollection(name string) *mongo.Collection {
	if DB == nil {
		log.Fatal("Database chưa được khởi tạo.")
	}
	return DB.Collection(name)
}
