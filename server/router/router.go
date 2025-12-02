// File router.go giúp điều hướng yêu cầu đến API cụ thể
package router

import (
	"database/sql"
	"fmt"
	"log"
	"note_sharing_application/server/handlers"
	"note_sharing_application/server/middlewares"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}

	handlers.DB = db

	api := r.Group("/api")
	{
		api.POST("/register", handlers.RegisterHandler)
		api.POST("/login", handlers.LoginHandler)
		api.GET("/server-public-key-rsa", handlers.GetServerPublicKeyRSA)

		noteRoutes := api.Group("/notes")
		noteRoutes.Use(middlewares.AuthMiddleware())
		{
			noteRoutes.POST("", handlers.CreateNoteHandler)
			noteRoutes.GET("", handlers.GetNoteHandler)
		}
	}
	return r
}
