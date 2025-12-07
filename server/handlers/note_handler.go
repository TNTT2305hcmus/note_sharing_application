package handlers

import (
	"net/http"
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
	notes, err := services.ViewOwnedNotes(c.Request.Context(), ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notes: " + err.Error()})
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
	urls, err := services.ViewReceivedNoteURLs(c.Request.Context(), receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch received URLs: " + err.Error()})
		return
	}
	if urls == nil {
		c.JSON(http.StatusOK, []interface{}{})
	} else {
		c.JSON(http.StatusOK, urls)
	}
}

func DeleteNoteHandler(c *gin.Context) {
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

	err := services.DeleteNote(c.Request.Context(), noteID, ownerID)

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
