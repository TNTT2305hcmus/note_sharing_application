package handlers

import (
	"note_sharing_application/server/services"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	NoteService *services.NoteService
}

func NewNoteHandler(s *services.NoteService) *NoteHandler {
	return &NoteHandler{
		NoteService: s,
	}
}

// lấy tất cả các notes do user hiện tại tạo
func (h *NoteHandler) GetOwnedNotes(c *gin.Context) {
	// userID cho khớp với auth_middleware.go
	currentUserID := c.GetString("userId")

	if currentUserID == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// gọi service và gửi kết quả cho client
	notes, err := h.NoteService.ViewOwnedNotes(c.Request.Context(), currentUserID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": notes})
}

// lấy tất cả các notes được gửi đến user hiện tại
func (h *NoteHandler) GetReceivedNotes(c *gin.Context) {
	currentUserID := c.GetString("userId")
	if currentUserID == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// gọi service và gửi kết quả cho client
	notes, err := h.NoteService.ViewReceivedNotes(c.Request.Context(), currentUserID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": notes})
}

func (h *NoteHandler) DeleteNote(c *gin.Context) {

	// lấy id từ param
	noteID := c.Param("note_id")

	// lấy userID từ context, nhớ khớp với auth_middleware.go
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// gọi service
	err := h.NoteService.DeleteNote(c.Request.Context(), noteID, userID)

	// gửi kết quả cho client
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "bạn không có quyền xóa ghi chú này":
			c.JSON(403, gin.H{"error": errMsg})
		case "ghi chú không tồn tại":
			c.JSON(404, gin.H{"error": errMsg})
		case "Note ID không hợp lệ":
			c.JSON(400, gin.H{"error": errMsg})
		default:
			c.JSON(500, gin.H{"error": "Lỗi hệ thống: " + errMsg})
		}
		return
	}

	c.JSON(200, gin.H{"message": "Xóa ghi chú thành công"})
}
