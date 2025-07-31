package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"notedteam.backend/config"
	"notedteam.backend/models"
	"notedteam.backend/ws" // Impor package WebSocket kita

	"github.com/gin-gonic/gin"
)

// --- Struct untuk Input Data ---

// CreateTodoInput mendefinisikan data yang dibutuhkan untuk membuat todo baru.
type CreateTodoInput struct {
	Title       string             `json:"title" binding:"required"`
	Description string             `json:"description"`
	Urgency     models.UrgencyType `json:"urgency"`
	DueDate     *time.Time         `json:"due_date"`
}

// UpdateTodoInput mendefinisikan data yang bisa diubah pada sebuah todo.
// Menggunakan pointer agar field yang tidak diisi (kosong/null) tidak ikut di-update.
type UpdateTodoInput struct {
	Title       *string             `json:"title"`
	Description *string             `json:"description"`
	Status      *models.StatusType  `json:"status"`
	Urgency     *models.UrgencyType `json:"urgency"`
	DueDate     *time.Time          `json:"due_date"`
}

// --- Fungsi Controller ---

// CreateTodo membuat sebuah todo baru di dalam sebuah tim.
// Rute: POST /api/teams/:teamId/todos
func CreateTodo(c *gin.Context) {
	teamIdStr := c.Param("teamId")

	teamId, err := strconv.ParseUint(teamIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Team ID format"})
		return
	}

	var input CreateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creatorID, _ := c.Get("user_id")
	// --- PERUBAHAN DI SINI ---
	urgency := input.Urgency
	if urgency == "" { // Jika klien tidak mengirim urgensi, gunakan default 'low'
		urgency = models.UrgencyLow
	}
	todo := models.Todo{
		Title:       input.Title,
		Description: input.Description,
		Status:      models.StatusPending,
		Urgency:     models.UrgencyLow,
		DueDate:     input.DueDate,
		TeamID:      uint(teamId),
		CreatorID:   creatorID.(uint),
		EditorID:    creatorID.(uint),
	}
	config.DB.Preload("Creator").Preload("Editor").First(&todo, todo.ID)

	if err := config.DB.Create(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	// --- Integrasi WebSocket ---
	msg := ws.Message{Event: "todo_created", Data: todo}
	jsonMsg, _ := json.Marshal(msg)
	ws.AppHub.Broadcast <- struct {
		Message []byte
		TeamID  uint
	}{Message: jsonMsg, TeamID: todo.TeamID}
	// -------------------------

	c.JSON(http.StatusCreated, gin.H{"data": todo})
}

// GetTeamTodos mengambil semua todo dari sebuah tim.
// Rute: GET /api/teams/:teamId/todos
func GetTeamTodos(c *gin.Context) {
	teamID := c.Param("teamId")
	var todos []models.Todo

	if err := config.DB.Preload("Creator").Preload("Editor").Where("team_id = ?", teamID).Order("created_at desc").Find(&todos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todos})
}

// UpdateTodo memperbarui sebuah todo yang spesifik.
// Rute: PUT /api/teams/:teamId/todos/:todoId
func UpdateTodo(c *gin.Context) {
	teamId := c.Param("teamId")
	todoId := c.Param("todoId")
	editorID, _ := c.Get("user_id")
	var input UpdateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var todo models.Todo
	if err := config.DB.Where("id = ? AND team_id = ?", todoId, teamId).First(&todo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found in this team"})
		return
	}
	todo.EditorID = editorID.(uint)
	if err := config.DB.Model(&todo).Updates(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}
	config.DB.Preload("Creator").Preload("Editor").First(&todo, todo.ID)

	// --- Integrasi WebSocket ---
	msg := ws.Message{Event: "todo_updated", Data: todo}
	jsonMsg, _ := json.Marshal(msg)
	ws.AppHub.Broadcast <- struct {
		Message []byte
		TeamID  uint
	}{Message: jsonMsg, TeamID: todo.TeamID}
	// -------------------------

	c.JSON(http.StatusOK, gin.H{"data": todo})
}

// DeleteTodo menghapus sebuah todo.
// Rute: DELETE /api/teams/:teamId/todos/:todoId
func DeleteTodo(c *gin.Context) {
	teamIdStr := c.Param("teamId")
	teamId, _ := strconv.ParseUint(teamIdStr, 10, 32)
	todoId := c.Param("todoId")

	var todo models.Todo
	if err := config.DB.Where("id = ? AND team_id = ?", todoId, uint(teamId)).First(&todo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found in this team"})
		return
	}

	if err := config.DB.Delete(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	// --- Integrasi WebSocket ---
	// Kirim hanya ID dari todo yang dihapus
	msg := ws.Message{Event: "todo_deleted", Data: gin.H{"id": todo.ID}}
	jsonMsg, _ := json.Marshal(msg)
	ws.AppHub.Broadcast <- struct {
		Message []byte
		TeamID  uint
	}{Message: jsonMsg, TeamID: uint(teamId)}
	// -------------------------

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
