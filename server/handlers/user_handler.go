package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// API trả về public_key của 1 client
// GET /api/users/:username/pubkey
func GetUserPublicKey(c *gin.Context) {
	// Lấy username từ URL param
	targetUsername := c.Param("username")

	var publicKey string

	// Query DB để lấy PublicKey của user đó
	err := DB.QueryRow("SELECT PublicKey FROM Users WHERE Username = ?", targetUsername).Scan(&publicKey)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Trả về JSON chuẩn
	c.JSON(http.StatusOK, gin.H{
		"username":   targetUsername,
		"public_key": publicKey,
	})
}
