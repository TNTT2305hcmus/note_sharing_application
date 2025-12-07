package handlers

import (
	"net/http"
	"note_sharing_application/server/models"
	"note_sharing_application/server/services"

	"github.com/gin-gonic/gin"
)

// lấy tất cả các notes do user hiện tại tạo
func GetOwnedNotes(c *gin.Context) {
	// userID cho khớp với auth_middleware.go
	ownerID := c.GetString("user_id")
	if ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID not found"})
		return
	}

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

// lấy tất cả các notes được gửi đến user hiện tại
func GetReceivedNoteURLs(c *gin.Context) {
	receiverID := c.GetString("user_id")
	if receiverID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	urls, err := services.ViewReceivedNoteURLs(receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if urls == nil {
		c.JSON(http.StatusOK, []interface{}{})
	} else {
		c.JSON(http.StatusOK, urls)
	}
}

func DeleteNote(c *gin.Context) {
	noteID := c.Param("id")
	if noteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Note ID is required"})
		return
	}

	ownerID := c.GetString("user_id")
	if ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := services.DeleteNote(noteID, ownerID)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "invalid note ID format" {
			// Lỗi 400: ID gửi lên không đúng định dạng Hex
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		} else if errMsg == "non-exist note id" {
			// Lỗi 404: Không tìm thấy note hoặc không có quyền (trả về 404 để bảo mật)
			c.JSON(http.StatusNotFound, gin.H{"error": "Note not found or access denied"})
		} else {
			// Lỗi 500: Lỗi DB hoặc hệ thống khác
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
		return
	}

	// Bước 5: Phản hồi thành công
	c.JSON(http.StatusOK, gin.H{"message": "Note and related URLs deleted successfully"})
}

func CreateNote(c *gin.Context) {
	ownerIDStr := c.GetString("user_id")
	if ownerIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	noteID, err := services.CreateNote(req.Title, req.CipherText, req.EncryptedAesKey, req.OwnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Note created successfully",
		"note_id": noteID,
	})

}

func DeleteSharedNote(c *gin.Context) {
	noteID := c.Param("note_id")
	ownerID := c.GetString("user_id")

	if ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := services.DeleteSharedNote(noteID, ownerID)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "invalid note ID format" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		} else if errMsg == "cannot revoke: note not found, unauthorized, or it is an original note" {
			// Trả về 403 (Forbidden) hoặc 404 tùy ý định, ở đây 400/403 để báo user biết họ đang cố xóa cái không được xóa
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delete sharing successfully"})
}
