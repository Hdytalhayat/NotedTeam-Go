// controllers/todo_controller.go
package controllers

import (
	"net/http"

	"notedteam.backend/config"
	"notedteam.backend/models"

	"github.com/gin-gonic/gin"
)

type CreateTodoInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

func CreateTodo(c *gin.Context) {
	var input CreateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id") // Ambil user_id dari middleware

	todo := models.Todo{
		Title:       input.Title,
		Description: input.Description,
		Status:      models.StatusPending, // Status default saat dibuat
		Urgency:     models.UrgencyLow,    // Urgensi default
		UserID:      userID.(uint),
	}

	if err := config.DB.Create(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": todo})
}

func GetTodos(c *gin.Context) {
	var todos []models.Todo
	userID, _ := c.Get("user_id")

	// Cari semua todo yang UserID-nya cocok dengan user yang sedang login
	config.DB.Where("user_id = ?", userID).Find(&todos)

	c.JSON(http.StatusOK, gin.H{"data": todos})
}

type UpdateTodoInput struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      models.StatusType  `json:"status"`
	Urgency     models.UrgencyType `json:"urgency"`
}

func UpdateTodo(c *gin.Context) {
	var todo models.Todo
	userID, _ := c.Get("user_id")

	// Pastikan todo ada dan dimiliki oleh user yang login
	if err := config.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&todo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	var input UpdateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&todo).Updates(input)
	c.JSON(http.StatusOK, gin.H{"data": todo})
}

func DeleteTodo(c *gin.Context) {
	var todo models.Todo
	userID, _ := c.Get("user_id")

	// Pastikan todo ada dan dimiliki oleh user yang login
	if err := config.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&todo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	config.DB.Delete(&todo)
	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
