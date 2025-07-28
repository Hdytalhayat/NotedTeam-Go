// main.go
package main

import (
	"log"

	"notedteam.backend/config"
	"notedteam.backend/controllers"
	"notedteam.backend/middlewares" // Impor package middleware
	"notedteam.backend/models"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	config.ConnectDatabase()

	// Tambahkan model baru ke migrasi
	err := config.DB.AutoMigrate(&models.User{}, &models.Todo{}, &models.Team{})
	if err != nil {
		log.Fatal("failed to migrate database")
	}

	// Grup rute Auth (tidak berubah)
	public := r.Group("/auth")
	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)

	// Rute yang diproteksi untuk API
	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		// Rute untuk manajemen tim
		protected.POST("/teams", controllers.CreateTeam)
		protected.GET("/teams", controllers.GetMyTeams)

		// Grup rute untuk aksi di dalam sebuah tim spesifik
		// Semua rute di sini memerlukan user untuk menjadi anggota tim
		teamRoutes := protected.Group("/teams/:teamId")
		teamRoutes.Use(middlewares.TeamMemberMiddleware())
		{
			// Invite
			teamRoutes.POST("/invite", controllers.InviteUserToTeam)

			// CRUD untuk Todos di dalam tim
			teamRoutes.POST("/todos", controllers.CreateTodo) // Sesuaikan nama fungsi controllernya
			teamRoutes.GET("/todos", controllers.GetTeamTodos)
			teamRoutes.PUT("/todos/:todoId", controllers.UpdateTodo)
			teamRoutes.DELETE("/todos/:todoId", controllers.DeleteTodo)
		}
	}

	log.Println("Starting server on :8080")
	r.Run(":8080")
}
