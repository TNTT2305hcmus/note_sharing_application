package configs

import (
	"context"
	"log"
	"note_sharing_application/server/models"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB(dbName string) {
	// Lấy biến môi trường trong .env
	mongoURI := os.Getenv("MONGO_URI")

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
	log.Println("Connected to MongoDB successfully")

	DB = client.Database(dbName)

	// Lấy collection cần tạo Index ("urls")
	urlsCollection := DB.Collection("urls")

	// Gọi hàm tạo Index với đầy đủ tham số
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
	if DB == nil {
		log.Fatal("Database chưa được khởi tạo.")
	}
	return DB.Collection(name)
}
