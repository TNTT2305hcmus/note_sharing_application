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
			// API yêu cầu đăng ký
			auth.POST("/register", handlers.RegisterHandler)
			// API yêu cầu đăng nhập
			auth.POST("/login", handlers.LoginHandler)
			// API yêu cầu lấy pubKey RSA của server
			auth.GET("/server-public-key-rsa", handlers.GetServerPublicKeyRSA)
			// API yêu cầu lấy pubKey của client khác
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
				noteRoutes.POST("", middlewares.ValidateCreateNote(), handlers.CreateNote)

				// DELETE /notes/:note_id
				noteRoutes.DELETE("/:note_id", middlewares.ValidateDeleteNote(), handlers.DeleteNote)

				// DELETE /notes/shared/:note_id
				noteRoutes.DELETE("/shared/:note_id", middlewares.ValidateDeleteSharedNote(), handlers.DeleteSharedNote)

				// GET /notes/owned
				noteRoutes.GET("/owned", middlewares.ValidateGetOwnedNotes(), handlers.GetOwnedNotes)

				// GET /notes/received
				noteRoutes.GET("/received", middlewares.ValidateGetReceivedNoteURLs(), handlers.GetReceivedNoteURLs)

				// --- URL SHARING ROUTES (Mới thêm vào) ---

				// Tạo URL chia sẻ cho một Note cụ thể
				// Có thêm middleware: ValidateCreateUrl (check chủ sở hữu, check metadata)
				noteRoutes.POST("/:note_id/url", middlewares.ValidateCreateUrl(), handlers.CreateNoteUrl)

				//Nếu muốn xem thì cần tìm 1 url sẵn trước thì mới được truy cập
				noteRoutes.GET("/:note_id/url", middlewares.ValidateUrlAccess(), handlers.GetNoteUrl)
			}
			api.GET("/note/:url_id", middlewares.ValidateUrlAccess(), middlewares.ValidateUrl(), handlers.ViewNoteHandler)
		}
	}
	return r
}