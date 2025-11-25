package handlers

import (
	"fmt"
	// "net/http"
	"github.com/gin-gonic/gin"
)

func CreateNoteHandler(c *gin.Context) {
	fmt.Println("CreateNoteHandler() is running...")

}

func GetNoteHandler(c *gin.Context) {
	fmt.Println("GetNoteHandler() is running...")
}
