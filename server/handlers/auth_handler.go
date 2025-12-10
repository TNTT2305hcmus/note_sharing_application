package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"note_sharing_application/server/models"
	"note_sharing_application/server/services"
	"note_sharing_application/server/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection *mongo.Collection

// API register new account
func RegisterHandler(c *gin.Context) {
	fmt.Println("RegisterHandler() is running...")

	// Parse
	var req models.RegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	// Decrypted Password (RSA/OAEP) from client
	encryptedPassBytes, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": "EncyptedPassword is not Base64"})
		return
	}

	rawPassword, err := utils.DecryptOAEP(encryptedPassBytes)
	if err != nil {
		c.JSON(400, gin.H{"error": "Error decypted encryptedPassword"})
		return
	}

	// Check user is valid
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := UserCollection.CountDocuments(ctx, bson.M{"username": req.Username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Generate Salt and hash password
	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fail to generate salt"})
		return
	}
	hashPassword := utils.HashPassword(rawPassword, salt)

	// Create Object User
	newUser := models.User{
		Username:          req.Username,
		EncryptedPassword: hashPassword,
		Salt:              salt,
		EncryptedPrivKey:  req.EncryptedPrivKey,
		PubKey:            req.PublicKey,
	}

	// Insert to DB
	_, err = UserCollection.InsertOne(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database insert failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully Registered",
	})
}

// API login
func LoginHandler(c *gin.Context) {
	fmt.Println("LoginHandler() is running...")

	var req models.LoginRequest

	// Parse
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	// Decrypted password (RSA/OAEP) from client
	encryptedPassBytes, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": "EncyptedPassword is not Base64"})
		return
	}

	rawPassword, err := utils.DecryptOAEP(encryptedPassBytes)
	if err != nil {
		c.JSON(400, gin.H{"error": "Error decypted encryptedPassword"})
		return
	}

	// Check user in db
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var foundUser models.User
	err = UserCollection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&foundUser)

	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect username or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database select failed"})
		return
	}

	// Check hashpassword
	match := utils.CheckPasswordHash(rawPassword, foundUser.Salt, foundUser.EncryptedPassword)

	if !match {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect username or password"})
		return
	}

	// Generate JWT token
	// ObjectID -> Hex string
	tokenString, err := services.GenerateAuthJWT(foundUser.ID.Hex(), foundUser.Username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fail to generate JWT Token"})
		return
	}

	// 6. Return response
	c.JSON(http.StatusOK, gin.H{
		"message":           "Succesfully Login",
		"token":             tokenString,
		"encrypted_privKey": foundUser.EncryptedPrivKey,
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
