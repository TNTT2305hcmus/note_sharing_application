package middlewares

import (
	"context"
	"net/http"
	"note_sharing_application/server/configs"
	"note_sharing_application/server/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ValidateCreateUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		noteId := c.Param("note_id")

		//Qua auth_middleware nên có username trong context chỉ cần lấy ra
		currentUser := c.GetString("userId")

		// 1. Kiểm tra Note có tồn tại không
		var note models.Note
		id, _ := primitive.ObjectIDFromHex(noteId)
		coll := configs.GetCollection("notes")
		err := coll.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&note)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Note không tồn tại"})
			return
		}

		// 2. Kiểm tra người yêu cầu có phải chủ sở hữu không
		if note.OwnerID != currentUser {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Yêu cầu không hợp lệ"})
			return
		}

		//Lấy metadata nằm trong body của request
		var req models.CreateUrlRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Thông tin không hợp lệ"})
			return
		}

		if req.MaxAccess <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Số lượt truy cập tối đa phải > 0"})
			return
		}

		// Parse thử thời gian
		if _, err := time.ParseDuration(req.ExpiresIn); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Định dạng thời gian sai (vd: 1h, 30m)"})
			return
		}

		// Lưu thông tin đã parse vào Context để Handler dùng
		c.Set("expires_in", req.ExpiresIn)
		c.Set("max_access", req.MaxAccess)
		c.Set("shared_encrypted_aes_key", req.SharedEncryptedAESKey)
		c.Set("receiver", req.Receiver)

		c.Next()
	}
}

// Kiểm tra quyền truy cập có được lấy url hay không
func ValidateNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		noteId := c.Param("note_id")

		var note models.Note
		id, _ := primitive.ObjectIDFromHex(noteId)
		noteColl := configs.GetCollection("notes")
		err := noteColl.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&note)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Note không tồn tại"})
			return
		}

		c.Next()
	}
}

// Nếu đã có quyền truy cập thì kiểm tra Url có tồn tại
func ValidateUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		urlId := c.Param("url_id")
		receiver := c.GetString("username")

		var url models.Url
		id, _ := primitive.ObjectIDFromHex(urlId)
		coll := configs.GetCollection("urls")
		err := coll.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&url)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Liên kết sai hoặc đã hết hạn"})
			return
		}

		if receiver != url.Receiver {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Yêu cầu không hợp lệ"})
			return
		}

		c.Set("url", url)
		c.Next()
	}
}
