// File router.go giúp điều hướng yêu cầu đến API cụ thể
package router

import (
	"note_sharing_application/server/handlers"
	"github.com/gin-gonic/gin"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	

	db, err := sql.Open("mysql", "root:thaiHCMUS@2023@/note_sharing_application")
	if err != nil {
		log.Fatal(err)
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
