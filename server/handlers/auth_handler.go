package handlers

import (
	"encoding/base64"
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

// API đăng ký tài khoản mới
func RegisterHandler(c *gin.Context) {
	fmt.Println("RegisterHandler() is running...")

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

	// Chuyển encrytedPassword được gửi ở client dưới dạng byte thành dạng Base64
	encryptedPassBytes, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": "EncyptedPassword is not Base64"})
		return
	}

	// Giải mã password để tiến hành hash và lưu vào db
	rawPassword, err := utils.DecryptOAEP(encryptedPassBytes)
	if err != nil {
		c.JSON(400, gin.H{"error": "Error decypted encryptedPassword"})
		return
	}

	// GenerateSalt
	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Fail to generate salt",
		})
	}
	// Hash password for saving to database
	hashPassword := utils.HashPassword(rawPassword, salt)

	// Luu DB
	_, err = DB.Exec(
		"INSERT INTO Users (Username, PasswordHash, Salt, EncryptedPrivateKey, PublicKey) VALUES (?, ?, ?, ?, ?)",
		req.Username, hashPassword, salt, req.EncryptedPrivateKey, req.PublicKey,
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

// API đăng nhập
func LoginHandler(c *gin.Context) {
	fmt.Println("LoginHandler() is running...")

	var req models.LoginRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Chuyển encrytedPassword được gửi ở client dưới dạng byte thành dạng Base64
	encryptedPassBytes, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": "EncyptedPassword is not Base64"})
		return
	}

	// Giải mã password để tiến hành hash và lưu vào db
	rawPassword, err := utils.DecryptOAEP(encryptedPassBytes)
	if err != nil {
		c.JSON(400, gin.H{"error": "Error decypted encryptedPassword"})
		return
	}

	var storedID int
	var storedPassHash string
	var storedSalt string
	var storedEncryptedPrivKey string

	err = DB.QueryRow(
		"SELECT ID, PasswordHash, Salt, EncryptedPrivateKey FROM Users WHERE Username = ?",
		req.Username,
	).Scan(&storedID, &storedPassHash, &storedSalt, &storedEncryptedPrivKey)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "SQL Error",
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
	match := utils.CheckPasswordHash(rawPassword, storedSalt, storedPassHash)

	if !match {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unsuccesfully Login",
		})
		return
	}

	// Generate JWT token from services
	// Lưu ý: ID trong DB là int, hàm GenerateJWT cần string, nên convert
	tokenString, err := services.GenerateJWT(strconv.Itoa(storedID), req.Username)

	fmt.Println("\nToken String được sinh ra: \n", tokenString)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Fail to generate JWT Token",
		})
		return
	}

	// Return to client (message + JWT Token + EncryptedPrivKey)
	c.JSON(http.StatusOK, gin.H{
		"message":               "Succesfully Login",
		"token":                 tokenString,
		"encrypted_private_key": storedEncryptedPrivKey,
	})

}

// API get server public key RSA
func GetServerPublicKeyRSA(c *gin.Context) {
	pemString, err := utils.ExportPublicKeyAsPEM()
	if err != nil {
		c.JSON(500, gin.H{"error": "Error export Server public key RSA"})
		return
	}
	c.JSON(200, gin.H{"server-public-key-rsa": pemString})
}
