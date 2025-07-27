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

	// Koneksi ke database
	config.ConnectDatabase()

	// Migrasi otomatis model User dan Todo
	err := config.DB.AutoMigrate(&models.User{}, &models.Todo{}) // Tambahkan &models.Todo{}
	if err != nil {
		log.Fatal("failed to migrate database")
	}

	// Rute publik untuk autentikasi
	public := r.Group("/auth")
	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)

	// Rute yang diproteksi untuk API
	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware()) // Gunakan middleware otentikasi
	{
		protected.POST("/todos", controllers.CreateTodo)
		protected.GET("/todos", controllers.GetTodos)
		protected.PUT("/todos/:id", controllers.UpdateTodo)
		protected.DELETE("/todos/:id", controllers.DeleteTodo)
	}

	// Jalankan server
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("failed to run server")
	}
}
