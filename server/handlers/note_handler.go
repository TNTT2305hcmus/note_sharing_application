package handlers

import (
	"net/http"
	"note_sharing_application/server/models"
	"note_sharing_application/server/services"

	"github.com/gin-gonic/gin"
)

func CreateNote(c *gin.Context) {
	// Lấy dữ liệu đã validate từ Context
	// Vì c.Get trả về interface{}, ta cần "ép kiểu" (Type Assertion) về đúng struct
	reqVal, exists := c.Get("validatedRequest")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống: Mất dữ liệu request trong context"})
		return
	}

	// Ép kiểu interface{} -> models.CreateNoteRequest
	req := reqVal.(models.CreateNoteRequest)

	// Gọi Service
	// Lưu ý: Lúc này req.OwnerID chắc chắn là ID của người đang đăng nhập
	ownerID := c.GetString("userId")
	noteID, err := services.CreateNote(req.CipherText, req.EncryptedAesKey, ownerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo ghi chú: " + err.Error()})
		return
	}

	// Phản hồi thành công
	c.JSON(http.StatusCreated, gin.H{
		"note_id": noteID,
	})
}

// lấy tất cả các notes do user hiện tại tạo
func GetOwnedNotes(c *gin.Context) {
	// userID cho khớp với auth_middleware.go
	ownerID := c.GetString("userId")

	// gọi service và gửi kết quả cho client
	notes, err := services.ViewOwnedNotes(ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if notes == nil {
		c.JSON(http.StatusOK, []interface{}{})
	} else {
		c.JSON(http.StatusOK, notes)
	}
}

// lấy tất cả các URLs được gửi đến user hiện tại
func GetReceivedNoteURLs(c *gin.Context) {
	receiver := c.GetString("username")

	urls, err := services.ViewReceivedNoteURLs(receiver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if urls == nil {
		c.JSON(http.StatusOK, []interface{}{})
	} else {
		res := []models.UrlResponse{}

		for _, url := range urls {
			var item models.UrlResponse
			item.ID = "localhost:8080/note/" + url.ID.Hex()
			item.NoteID = url.NoteID
			item.SenderID = url.Sender
			item.ReceiverID = url.Receiver
			item.ExpiresAt = url.ExpiresAt
			item.MaxAccess = url.MaxAccess
			res = append(res, item)
		}
		c.JSON(http.StatusOK, res)
	}
}

func DeleteNote(c *gin.Context) {
	// 1. Lấy thông tin (Đã được kiểm chứng an toàn 100% bởi Middleware)
	noteId := c.Param("note_id")

	// 2. Gọi Service để thực hiện xóa
	// Lúc này Service không cần kiểm tra quyền sở hữu nữa, chỉ cần thực hiện lệnh Delete
	err := services.DeleteNote(noteId)

	if err != nil {
		// Vì Middleware đã check tồn tại, lỗi ở đây thường là lỗi hệ thống (DB down, transaction fail...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa note: " + err.Error()})
		return
	}

	// 3. Phản hồi thành công
	c.JSON(http.StatusOK, gin.H{"message": "Xóa Note và các dữ liệu liên quan thành công"})
}

func DeleteSharedNote(c *gin.Context) {
	// 1. Lấy dữ liệu (An toàn tuyệt đối nhờ Middleware)
	noteID := c.Param("note_id")
	owner := c.GetString("username") // Dùng key "userId"

	// 2. Gọi Service
	// Service lúc này chỉ cần thực hiện logic: "Xóa tất cả URL share có note_id = X và owner_id = Y"
	err := services.DeleteSharedNote(noteID, owner)

	if err != nil {
		// Middleware đã check tồn tại note, nên lỗi ở đây thường là lỗi Server/DB
		// hoặc logic nghiệp vụ đặc thù (ví dụ: note này chưa từng được share)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể hủy chia sẻ: " + err.Error()})
		return
	}

	// 3. Phản hồi thành công
	c.JSON(http.StatusOK, gin.H{"message": "Đã hủy chia sẻ thành công (Revoked)"})
}
