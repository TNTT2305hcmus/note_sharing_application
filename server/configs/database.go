package configs

import (
	"context"
	"log"
	"note_sharing_application/server/models"
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

	// 1. Gán vào biến global DB
	DB = client.Database(os.Getenv("DB_NAME"))

	// 2. Lấy đúng collection cần tạo Index ("urls")
	urlsCollection := DB.Collection("urls")

	// 3. Gọi hàm tạo Index với đầy đủ tham số
	// Truyền context và collection vào
	err = models.CreateTTLIndex(context.Background(), urlsCollection)
	if err != nil {
		// Có thể log warning hoặc fatal tùy mức độ nghiêm trọng bạn muốn
		log.Printf("Cảnh báo: Không thể tạo TTL Index cho urls: %v", err)
	} else {
		log.Println("Đã kích hoạt tính năng tự xóa (TTL Index) cho urls")
	}
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}
