// main.go
package main

import (
	"log"

	"NotedTeam/config"
	"NotedTeam/controllers"
	"NotedTeam/models"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Koneksi ke database
	config.ConnectDatabase()

	// Migrasi otomatis model User
	err := config.DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("failed to migrate database")
	}

	// Rute publik untuk autentikasi
	public := r.Group("/auth")
	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)

	// Contoh rute yang akan diproteksi nanti
	// protected := r.Group("/api")
	// protected.Use(middlewares.AuthMiddleware()) // Ini akan ditambahkan nanti
	// protected.GET("/profile", func(c *gin.Context) {
	//     c.JSON(http.StatusOK, gin.H{"message": "This is a protected route"})
	// })

	// Jalankan server
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("failed to run server")
	}
}
