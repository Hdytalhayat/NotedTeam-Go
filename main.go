package main

import (
	"log"

	"notedteam.backend/config"
	"notedteam.backend/controllers"
	"notedteam.backend/middlewares"
	"notedteam.backend/models"
	"notedteam.backend/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	config.ConnectDatabase()

	log.Println("Running database migrations...")
	err := config.DB.AutoMigrate(&models.User{}, &models.Team{}, &models.Todo{}, &models.Invitation{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	go ws.AppHub.Run()
	log.Println("WebSocket Hub started.")

	// --- STRUKTUR RUTE YANG DIPERBAIKI ---

	// 1. Grup untuk rute publik (tanpa otentikasi)
	public := r.Group("/auth")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.GET("/verify", controllers.VerifyEmail)
		public.POST("/forgot-password", controllers.ForgotPassword)
		public.GET("/reset-password-page", controllers.ShowResetPasswordPage)
		public.POST("/reset-password", controllers.ResetPassword)
	}

	// 2. Grup untuk API standar yang menggunakan otentikasi via Header 'Authorization'
	api := r.Group("/api")
	api.Use(middlewares.AuthMiddleware())
	{
		api.POST("/teams", controllers.CreateTeam)
		api.GET("/teams", controllers.GetMyTeams)

		teamRoutes := api.Group("/teams/:teamId")

		teamRoutes.Use(middlewares.TeamMemberMiddleware())
		{
			// Rute yang hanya butuh keanggotaan
			memberRoutes := teamRoutes.Group("")
			memberRoutes.Use(middlewares.TeamMemberMiddleware())
			memberRoutes.GET("", controllers.GetTeamDetails)
			memberRoutes.POST("/invite", controllers.InviteUserToTeam)
			memberRoutes.POST("/todos", controllers.CreateTodo)
			teamRoutes.GET("/todos", controllers.GetTeamTodos)
			teamRoutes.PUT("/todos/:todoId", controllers.UpdateTodo)
			teamRoutes.DELETE("/todos/:todoId", controllers.DeleteTodo)
			ownerRoutes := teamRoutes.Group("")
			ownerRoutes.Use(middlewares.TeamOwnerMiddleware()) // Gunakan middleware baru
			ownerRoutes.PUT("", controllers.UpdateTeam)        // PUT ke /api/teams/:teamId
			ownerRoutes.DELETE("", controllers.DeleteTeam)     // DELETE ke /api/teams/:teamId
			// Rute baru untuk manajemen undangan
			api.GET("/invitations", controllers.GetMyInvitations)
			api.POST("/invitations/:invitationId/respond", controllers.RespondToInvitation)
		}

	}

	// 3. Grup TERPISAH khusus untuk WebSocket dengan middleware-nya sendiri
	//    Middleware ini bisa menerima token dari query parameter.
	wsApi := r.Group("/api/ws")
	wsApi.Use(middlewares.WsAuthMiddleware())
	{
		wsApi.GET("/teams/:teamId", controllers.ServeWs)
	}

	// Jalankan server
	log.Println("Starting server on http://localhost:8080")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
