package handlers

import (
	"fmt"
	"net/http"
	"note_sharing_application/server/models"
	"note_sharing_application/server/services"
	"note_sharing_application/server/utils"
	"strconv"

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

	// Parse dữ liệu nhận từ client
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Check
	fmt.Println("Username: ", req.Username)
	fmt.Println("Password: ", req.Password)
	fmt.Println("Pulic Key: ", req.PublicKey)

	// GenerateSalt
	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Fail to generate salt",
		})
	}
	// Hash password for saving to database
	hashPassword := utils.HashPassword(req.Password, salt)

	// Luu DB
	_, err = DB.Exec(
		"INSERT INTO Users (Username, PasswordHash, Salt, PublicKey) VALUES (?, ?, ?, ?)",
		req.Username, hashPassword, salt, req.PublicKey,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database insert failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully Registed",
	})
}

func LoginHandler(c *gin.Context) {
	fmt.Println("LoginHandler() is running...")

	c.JSON(http.StatusOK, gin.H{
		"message": "Server is processing login request...",
	})

	var req models.LoginRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}
	var storedID int
	var storedPassword string
	var storedSalt string

	err = DB.QueryRow(
		"SELECT ID, PasswordHash, Salt FROM Users WHERE Username = ?",
		req.Username,
	).Scan(&storedID, &storedPassword, &storedSalt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unsuccesfully Login",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database select failed",
		})
		return
	}

	// Check
	match := utils.CheckPasswordHash(req.Password, storedSalt, storedPassword)

	if !match {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unsuccesfully Login",
		})
		return
	}

	// Generate JWT token from services
	// Lưu ý: ID trong DB là int, hàm GenerateJWT cần string, nên convert
	tokenString, err := services.GenerateJWT(strconv.Itoa(storedID), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Fail to generate JWT Token",
		})
		return
	}

	// Return to client (message + JWT Token)
	c.JSON(http.StatusOK, gin.H{
		"message": "Succesfully Login",
		"token":   tokenString,
	})

}
