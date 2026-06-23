package routes

import (
	"pintree-backend/config"
	"pintree-backend/handlers"
	"pintree-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// 使用CORS中间件
	router.Use(middleware.CORS(cfg))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 提供快照文件静态服务
	router.Static("/snapshots", "./uploads/snapshots")

	// 提供头像文件静态服务
	router.Static("/avatars", "./uploads/avatars")

	// API路由组
	api := router.Group("/api")
	{
		// 认证路由（无需登录）
		authHandler := handlers.NewAuthHandler(cfg)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)

			// 需要认证的路由
			protected := auth.Use(middleware.AuthMiddleware(cfg))
			{
				protected.GET("/profile", authHandler.GetProfile)
				protected.POST("/avatar", authHandler.UploadAvatar)
				protected.DELETE("/avatar", authHandler.DeleteAvatar)
				protected.PUT("/password", authHandler.ChangePassword)
			}

			// 管理员路由
			admin := auth.Use(middleware.AuthMiddleware(cfg), middleware.AdminMiddleware())
			{
				admin.GET("/users", authHandler.ListUsers)
				admin.PUT("/users/:id", authHandler.UpdateUser)
				admin.DELETE("/users/:id", authHandler.DeleteUser)
			}
		}

		// 公开API（部分需要认证）
		collectionHandler := handlers.NewCollectionHandler()
		collections := api.Group("/collections")
		{
			// 使用可选认证中间件，允许未登录用户访问，但如果已登录则获取用户信息
			collections.Use(middleware.OptionalAuthMiddleware(cfg))
			collections.GET("", collectionHandler.GetCollections)
			collections.GET("/:id", collectionHandler.GetCollectionByID)
			collections.GET("/slug/:slug", collectionHandler.GetCollectionBySlug)

			// 需要认证的路由
			protected := collections.Use(middleware.AuthMiddleware(cfg))
			{
				protected.POST("", collectionHandler.CreateCollection)
				protected.PUT("/:id", collectionHandler.UpdateCollection)
				protected.DELETE("/:id", collectionHandler.DeleteCollection)
				protected.PUT("/batch/sort", collectionHandler.BatchUpdateSortOrders)
			}
		}

		// 书签API
		bookmarkHandler := handlers.NewBookmarkHandler(cfg)
		bookmarks := api.Group("/bookmarks")
		{
			// 使用可选认证中间件，允许未登录用户访问，但如果已登录则获取用户信息
			bookmarks.Use(middleware.OptionalAuthMiddleware(cfg))
			bookmarks.GET("", bookmarkHandler.GetBookmarks)
			bookmarks.GET("/search", bookmarkHandler.SearchBookmarks)
			bookmarks.GET("/:id", bookmarkHandler.GetBookmarkByID)

			// 需要认证的路由
			protected := bookmarks.Use(middleware.AuthMiddleware(cfg))
			{
				protected.POST("", bookmarkHandler.CreateBookmark)
				protected.PUT("/:id", bookmarkHandler.UpdateBookmark)
				protected.DELETE("/:id", bookmarkHandler.DeleteBookmark)
				protected.POST("/:id/snapshot", bookmarkHandler.GenerateSnapshot)
				protected.GET("/export", bookmarkHandler.ExportBookmarks)
				protected.GET("/export/html", bookmarkHandler.ExportBookmarksHTML)
				protected.POST("/import/html", bookmarkHandler.ImportBookmarksHTML)
				protected.POST("/batch/delete", bookmarkHandler.BatchDeleteBookmarks)
				protected.PUT("/batch/move", bookmarkHandler.BatchMoveBookmarks)
				protected.PUT("/batch/sort", bookmarkHandler.BatchUpdateSortOrders)
			}
		}

		// 文件夹API
		folderHandler := handlers.NewFolderHandler()
		folders := api.Group("/folders")
		{
			// 使用可选认证中间件，允许未登录用户访问，但如果已登录则获取用户信息
			folders.Use(middleware.OptionalAuthMiddleware(cfg))
			folders.GET("", folderHandler.GetFolders)
			folders.GET("/:id", folderHandler.GetFolderByID)

			// 需要认证的路由
			protected := folders.Use(middleware.AuthMiddleware(cfg))
			{
				protected.POST("", folderHandler.CreateFolder)
				protected.PUT("/:id", folderHandler.UpdateFolder)
				protected.DELETE("/:id", folderHandler.DeleteFolder)
				protected.PUT("/batch/sort", folderHandler.BatchUpdateSortOrders)
			}
		}

		// 设置API
		settingHandler := handlers.NewSettingHandler()
		settings := api.Group("/settings")
		{
			settings.GET("", settingHandler.GetSettings)
			settings.GET("/initSettings", settingHandler.InitSettings)
			settings.GET("/:key", settingHandler.GetSettingByKey)

			// 需要管理员权限
			protected := settings.Use(middleware.AuthMiddleware(cfg), middleware.AdminMiddleware())
			{
				protected.PUT("/:key", settingHandler.UpdateSetting)
			}
		}

		// 标签API
		tags := api.Group("/tags")
		{
			tags.GET("", func(c *gin.Context) {
				c.JSON(200, gin.H{"tags": []string{}})
			})
		}

		// 搜索API
		search := api.Group("/search")
		{
			search.GET("/bookmarks", bookmarkHandler.SearchBookmarks)
		}

		// 公开设置API（无需认证）
		publicSettings := api.Group("/settings/public")
		{
			publicSettings.GET("", settingHandler.GetPublicSettings)
		}
	}

	return router
}
