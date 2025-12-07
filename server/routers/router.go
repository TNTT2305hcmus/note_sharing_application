package routers

import (
	"github.com/gin-gonic/gin"

	// Import các package nội bộ
	"note_sharing_application/server/handlers"
	"note_sharing_application/server/middlewares"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// khởi tạo handlers ở đây để truyền vào các route
	noteHandler := handlers.NewNoteHandler(nil)
	// group gốc
	api := r.Group("/api")
	{
		// group xác thực
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterHandler)
			auth.POST("/login", handlers.LoginHandler)
			auth.GET("/server-public-key-rsa", handlers.GetServerPublicKeyRSA)
			auth.GET("/users/:username/pubkey", handlers.GetUserPublicKey)
		}

		// group cần đăng nhập
		protected := api.Group("/")
		protected.Use(middlewares.AuthMiddleware())
		{
			noteRoutes := protected.Group("/notes")
			{
				// POST /api/notes
				// noteRoutes.POST("", handlers.CreateNote)
				// GET /api/notes/owned
				noteRoutes.GET("/owned", noteHandler.GetOwnedNotes)
				// GET /api/notes/received
				noteRoutes.GET("/received", noteHandler.GetReceivedNotes)
				// DELETE /api/notes/:note_id
				noteRoutes.DELETE("/:note_id", noteHandler.DeleteNote)
				// POST /api/notes/:note_id
				noteRoutes.POST("/:note_id", handlers.CreateNoteUrl)
				// GET /api/:note_id/url
				noteRoutes.GET("/api/:note_id/url", handlers.GetNoteUrl)
				// GET /api/:note_id/url
				noteRoutes.GET("/note/:url_id", handlers.ViewNoteHandler)
			}
		}
	}
	return r
}
