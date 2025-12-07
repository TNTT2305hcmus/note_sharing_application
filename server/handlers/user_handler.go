package handlers

import (
	"context"
	"net/http"
	"time"

	"note_sharing_application/server/models" // Import package models của bạn

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// API trả về public_key của 1 client
// GET /api/users/:username/pubkey
func GetUserPublicKey(c *gin.Context) {
	// Lấy username từ URL param
	targetUsername := c.Param("username")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var foundUser models.User

	err := UserCollection.FindOne(ctx, bson.M{"username": targetUsername}).Decode(&foundUser)
	// Not found user
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Other fails
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Return JSON
	c.JSON(http.StatusOK, gin.H{
		"username":   foundUser.Username,
		"public_key": foundUser.PubKey,
	})
}
