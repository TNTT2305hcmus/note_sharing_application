package middlewares

import (
	"context"
	"net/http"
	"note_sharing_application/server/configs"
	"note_sharing_application/server/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ValidateGetOwnedNotes() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Kiểm tra tính tồn tại của định danh người dùng (User ID)
		userID := c.GetString("userId")

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: User ID not found in context",
			})
			return
		}
		c.Next()
	}
}

func ValidateGetReceivedNoteURLs() gin.HandlerFunc {
	return func(c *gin.Context) {

		receiverID := c.GetString("userId")
		if receiverID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Receiver ID not found in context",
			})
			return
		}
		c.Next()
	}
}

func ValidateDeleteNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy Note ID từ URL
		noteIdHex := c.Param("note_id")

		// 2. Validate định dạng ObjectID (Fail Fast)
		// Nếu ID sai định dạng Hex, chặn ngay lập tức, không cần gọi DB
		id, err := primitive.ObjectIDFromHex(noteIdHex)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Note ID không hợp lệ"})
			return
		}

		// 3. Lấy User ID từ Context (Sửa lại key "userId" cho khớp AuthMiddleware)
		currentUserID := c.GetString("userId")
		if currentUserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Không tìm thấy thông tin người dùng"})
			return
		}

		// 4. Truy vấn Database để kiểm tra sự tồn tại và quyền sở hữu
		// Lưu ý: Việc truy vấn ở đây giúp Handler chính không cần lo về logic này nữa
		var note models.Note
		coll := configs.GetCollection("notes")

		// Tìm note theo _id
		err = coll.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&note)
		if err != nil {
			// Nếu lỗi là do không tìm thấy document
			if err == mongo.ErrNoDocuments {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Note không tồn tại"})
				return
			}
			// Lỗi hệ thống DB khác
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kết nối cơ sở dữ liệu"})
			return
		}

		// 5. Kiểm tra quyền sở hữu (Authorization)
		if note.OwnerID != currentUserID {
			// Trả về 403 Forbidden: Đã đăng nhập nhưng không có quyền truy cập tài nguyên này
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xóa Note này"})
			return
		}

		c.Next()
	}
}

func ValidateCreateNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Kiểm tra danh tính từ Context (Nguồn tin cậy nhất)
		userID := c.GetString("userId") // Key khớp với AuthMiddleware
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Không xác định được người dùng"})
			return
		}

		// 2. Parse dữ liệu từ JSON Body
		var req models.CreateNoteRequest
		// ShouldBindJSON sẽ kiểm tra các tag như `binding:"required"` trong struct của bạn
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ: " + err.Error()})
			return
		}

		// 5. Lưu object đã validate vào Context để Handler dùng
		// Key này dùng để truyền dữ liệu giữa Middleware và Handler
		c.Set("validatedRequest", req)

		c.Next()
	}
}

func ValidateDeleteSharedNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy Note ID từ URL (ví dụ: /notes/share/:note_id)
		noteIdHex := c.Param("note_id")

		// 2. Validate định dạng ObjectID (Fail Fast)
		// Chặn ngay nếu ID gửi lên không phải là chuỗi Hex 24 ký tự hợp lệ
		id, err := primitive.ObjectIDFromHex(noteIdHex)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Note ID không đúng định dạng"})
			return
		}

		// 3. Lấy User ID từ Context (Key "userId" từ AuthMiddleware)
		currentUserID := c.GetString("userId")
		if currentUserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Không xác định được danh tính"})
			return
		}

		// 4. Truy vấn Database để kiểm tra quyền sở hữu đối với GHI CHÚ GỐC
		// Logic: Để xóa link chia sẻ, bạn phải là chủ của ghi chú đó.
		var note models.Note
		coll := configs.GetCollection("notes")

		// Tìm note theo _id
		err = coll.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&note)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Ghi chú không tồn tại để thực hiện thao tác"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống cơ sở dữ liệu"})
			return
		}

		// 5. Kiểm tra quyền sở hữu (Authorization)
		if note.OwnerID != currentUserID {
			// Nếu người yêu cầu không phải chủ sở hữu -> Cấm (403 Forbidden)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thu hồi chia sẻ của ghi chú này"})
			return
		}

		// Nếu hợp lệ, cho phép đi tiếp vào Handler
		c.Next()
	}
}
