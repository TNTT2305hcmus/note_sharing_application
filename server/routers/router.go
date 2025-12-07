package routers

import (
	"github.com/gin-gonic/gin"

	// Import các package nội bộ
	"note_sharing_application/server/handlers"
	"note_sharing_application/server/middlewares"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	// group gốc
	api := r.Group("/")
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
			// Gom nhóm liên quan đến Notes: /api/notes
			noteRoutes := protected.Group("/notes")
			{
				// POST /notes
				// noteRoutes.POST("", handlers.CreateNote)

				// DELETE /notes/:note_id
				noteRoutes.DELETE("/:note_id", handlers.DeleteNoteHandler)

				// GET /notes/owned
				noteRoutes.GET("/owned", handlers.GetOwnedNotes)

				// GET /notes/inbox
				noteRoutes.GET("/received", handlers.GetReceivedNoteURLs)

				// --- URL SHARING ROUTES (Mới thêm vào) ---

				// 1. Tạo URL chia sẻ cho một Note cụ thể
				// Có thêm middleware: ValidateCreateUrl (check chủ sở hữu, check metadata)
				noteRoutes.POST("/:note_id/url", middlewares.ValidateCreateUrl(), handlers.CreateNoteUrl)

				//Nếu muốn xem thì cần tìm 1 url sẵn trước thì mới được truy cập
				noteRoutes.GET("/:note_id/url", middlewares.ValidateUrlAccess(), handlers.GetNoteUrl)

				//API xóa ghi chú, hủy chia sẻ viết ở đây

				// GET /notes/url_id
				noteRoutes.GET("/view/:url_id", handlers.ViewNoteHandler)
			}
			api.GET("/note/:url_id", middlewares.AuthMiddleware(), middlewares.ValidateUrlAccess(), middlewares.ValidateUrl(), handlers.ViewNoteHandler)
		}
	}
	return r
}
