package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"notedteam.backend/config"
	"notedteam.backend/models"
)

type CreateTeamInput struct {
	Name string `json:"name" binding:"required"`
}

func CreateTeam(c *gin.Context) {
	var input CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Ambil user ID dari context
	userID, _ := c.Get("user_id")

	// 2. Ambil objek user lengkap dari context dan lakukan type assertion
	userContext, exists := c.Get("user")
	if !exists {
		// Ini seharusnya tidak terjadi jika AuthMiddleware berjalan dengan benar
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User object not found in context"})
		return
	}

	// 3. Lakukan type assertion dari interface{} ke models.User
	user, ok := userContext.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object type in context"})
		return
	}

	// 4. Siapkan objek tim baru
	team := models.Team{
		Name:    input.Name,
		OwnerID: userID.(uint),
	}

	// 5. Gunakan Transaksi Database untuk memastikan integritas data
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Langkah A: Buat entitas tim di dalam transaksi
		if err := tx.Create(&team).Error; err != nil {
			// Jika gagal, kembalikan error untuk me-rollback transaksi
			return err
		}

		// Langkah B: Tambahkan user (pemilik) sebagai anggota pertama
		// Gunakan pointer (&user) untuk menambahkan asosiasi
		if err := tx.Model(&team).Association("Members").Append(&user); err != nil {
			// Jika gagal, kembalikan error untuk me-rollback transaksi
			return err
		}

		// Jika semua berhasil, kembalikan nil untuk meng-commit transaksi
		return nil
	})

	// Jika ada error selama transaksi, proses akan gagal
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team or add owner"})
		return
	}

	// Muat ulang data member agar tampil di response (opsional tapi bagus)
	config.DB.Preload("Members").First(&team, team.ID)

	c.JSON(http.StatusCreated, gin.H{"data": team})
}

type InviteInput struct {
	Email string `json:"email" binding:"required,email"`
}

func InviteUserToTeam(c *gin.Context) {
	var input InviteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teamId := c.Param("teamId")
	var userToInvite models.User
	var team models.Team

	// Cari user yang akan diundang berdasarkan email
	if err := config.DB.Where("email = ?", input.Email).First(&userToInvite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User to invite not found"})
		return
	}

	// Cari tim
	if err := config.DB.First(&team, teamId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Tambahkan user ke dalam tim
	if err := config.DB.Model(&team).Association("Members").Append(&userToInvite); err != nil {
		// Error ini bisa terjadi jika user sudah menjadi member
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member or failed to add"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully invited to the team"})
}

func GetMyTeams(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User

	// Cari user beserta tim yang diikutinya
	config.DB.Preload("Teams").First(&user, userID)

	c.JSON(http.StatusOK, gin.H{"data": user.Teams})
}
