package handlers

import (
	"fmt"
	"net/http"
	"note_sharing_application/server/models"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func RegisterHandler(c *gin.Context) {
	fmt.Println("RegisterHandler() is running...")

	c.JSON(http.StatusOK, gin.H{
		"message": "Server is processing register request...",
	})

	// struct luu thong tin nhan tu client
	var req models.RegisterRequest

	// Kiem tra tinh hop le du lieu va chuyen JSON -> struct
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}

	// In ra de check
	fmt.Println("Username: ", req.Username)
	fmt.Println("Password: ", req.Password)
	fmt.Println("Pulic Key: ", req.PublicKey)

	// Luu DB
	_, err = DB.Exec(
		"INSERT INTO Users (Username, PasswordHash, PublicKey) VALUES (?, ?, ?)",
		req.Username, req.Password, req.PublicKey,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "Database insert failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered"})

}

func LoginHandler(c *gin.Context) {
	fmt.Println("LoginHandler() is running...")

	c.JSON(http.StatusOK, gin.H{
		"message": "Server is processing login request...",
	})

	var req models.LoginRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}

	fmt.Println("Username: ", req.Username)
	fmt.Println("Password: ", req.Password)

	var existID int
	err = DB.QueryRow(
		"SELECT ID FROM Users WHERE Username = ? AND PasswordHash = ?",
		req.Username, req.Password,
	).Scan(&existID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"message": "Unsuccessful login"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "Database select failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User logined"})

}
