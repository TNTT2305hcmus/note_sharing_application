// File router.go giúp điều hướng yêu cầu đến API cụ thể
package router

import (
	"database/sql"
	"log"
	"note_sharing_application/server/handlers"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	db, err := sql.Open("mysql", "root:thaiHCMUS@2023@tcp(localhost:3306)/note_sharing_application")
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
		noteRoutes := api.Group("/notes")
		{
			noteRoutes.POST("", handlers.CreateNoteHandler)
			noteRoutes.GET("", handlers.GetNoteHandler)
		}
	}
	return r
}
