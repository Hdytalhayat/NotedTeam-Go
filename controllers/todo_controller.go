package controllers

import (
	"net/http"
	"strconv"

	"notedteam.backend/config"
	"notedteam.backend/models"

	"github.com/gin-gonic/gin"
)

// --- Struct untuk Input Data ---

// CreateTodoInput mendefinisikan data yang dibutuhkan untuk membuat todo baru.
type CreateTodoInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// UpdateTodoInput mendefinisikan data yang bisa diubah pada sebuah todo.
// Menggunakan pointer agar field yang tidak diisi (kosong/null) tidak ikut di-update.
type UpdateTodoInput struct {
	Title       *string             `json:"title"`
	Description *string             `json:"description"`
	Status      *models.StatusType  `json:"status"`
	Urgency     *models.UrgencyType `json:"urgency"`
}

// --- Fungsi Controller ---

// CreateTodo membuat sebuah todo baru di dalam sebuah tim.
// Rute: POST /api/teams/:teamId/todos
func CreateTodo(c *gin.Context) {
	// 1. Dapatkan ID Tim dari parameter URL
	teamIdStr := c.Param("teamId")
	teamId, err := strconv.ParseUint(teamIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Team ID"})
		return
	}

	// 2. Validasi input JSON
	var input CreateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Dapatkan ID pengguna yang membuat todo dari context (disediakan oleh AuthMiddleware)
	creatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	// 4. Buat objek Todo baru
	todo := models.Todo{
		Title:       input.Title,
		Description: input.Description,
		Status:      models.StatusPending, // Status default saat dibuat
		Urgency:     models.UrgencyLow,    // Urgensi default saat dibuat
		TeamID:      uint(teamId),
		CreatorID:   creatorID.(uint),
	}

	// 5. Simpan ke database
	if err := config.DB.Create(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	// 6. Kembalikan data todo yang baru dibuat
	c.JSON(http.StatusCreated, gin.H{"data": todo})
}

// GetTeamTodos mengambil semua todo dari sebuah tim.
// Rute: GET /api/teams/:teamId/todos
func GetTeamTodos(c *gin.Context) {
	// 1. Dapatkan ID Tim dari parameter URL
	teamId := c.Param("teamId")

	var todos []models.Todo

	// 2. Cari semua todo di database yang memiliki team_id yang cocok
	// Middleware sudah memastikan user adalah anggota tim ini, jadi kita tidak perlu cek lagi di sini.
	if err := config.DB.Where("team_id = ?", teamId).Order("created_at desc").Find(&todos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch todos"})
		return
	}

	// 3. Kembalikan daftar todo
	c.JSON(http.StatusOK, gin.H{"data": todos})
}

// UpdateTodo memperbarui sebuah todo yang spesifik.
// Rute: PUT /api/teams/:teamId/todos/:todoId
func UpdateTodo(c *gin.Context) {
	teamId := c.Param("teamId")
	todoId := c.Param("todoId")

	// 1. Validasi input JSON
	var input UpdateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Cari todo di database.
	// PENTING: Pastikan todo yang dicari (berdasarkan todoId) juga merupakan bagian dari tim yang benar (berdasarkan teamId).
	// Ini mencegah pengguna mengubah todo di tim lain secara tidak sengaja.
	var todo models.Todo
	if err := config.DB.Where("id = ? AND team_id = ?", todoId, teamId).First(&todo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found in this team"})
		return
	}

	// 3. Update model Todo dengan data dari input
	// Menggunakan DB.Model().Updates() memungkinkan pembaruan parsial.
	if err := config.DB.Model(&todo).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	// 4. Kembalikan data todo yang sudah diperbarui
	c.JSON(http.StatusOK, gin.H{"data": todo})
}

// DeleteTodo menghapus sebuah todo.
// Rute: DELETE /api/teams/:teamId/todos/:todoId
func DeleteTodo(c *gin.Context) {
	teamId := c.Param("teamId")
	todoId := c.Param("todoId")

	// 1. Cari todo di database.
	// Sama seperti Update, kita harus memastikan todo ini ada di dalam tim yang benar sebelum menghapusnya.
	var todo models.Todo
	if err := config.DB.Where("id = ? AND team_id = ?", todoId, teamId).First(&todo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found in this team"})
		return
	}

	// 2. Hapus todo dari database
	if err := config.DB.Delete(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	// 3. Kembalikan pesan sukses
	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
