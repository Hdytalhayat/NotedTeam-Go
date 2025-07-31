package controllers

import (
	"net/http"
	"strconv"

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

	// Konversi teamId string ke uint
	teamIdUint, err := strconv.ParseUint(teamId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var userToInvite models.User
	var team models.Team

	// Cari user yang akan diundang berdasarkan email
	if err := config.DB.Where("email = ?", input.Email).First(&userToInvite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User to invite not found"})
		return
	}

	// Cari tim berdasarkan ID
	if err := config.DB.First(&team, teamIdUint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Periksa apakah user sudah menjadi anggota
	var memberCount int64
	if err := config.DB.
		Table("team_members").
		Where("user_id = ? AND team_id = ?", userToInvite.ID, teamIdUint).
		Count(&memberCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check team membership"})
		return
	}
	if memberCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this team"})
		return
	}

	// Periksa apakah sudah ada undangan yang tertunda
	var existingInvitation models.Invitation
	if err := config.DB.
		Where("user_id = ? AND team_id = ? AND status = ?", userToInvite.ID, teamIdUint, models.InvitationPending).
		First(&existingInvitation).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "An invitation has already been sent to this user."})
		return
	}

	// Buat undangan baru
	invitation := models.Invitation{
		UserID: userToInvite.ID,
		TeamID: uint(teamIdUint),
		Status: models.InvitationPending,
	}

	if err := config.DB.Create(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User successfully invited to the team"})
}

func GetMyTeams(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User

	// Cari user beserta tim yang diikutinya
	config.DB.Preload("Teams").First(&user, userID)

	c.JSON(http.StatusOK, gin.H{"data": user.Teams})
}

type UpdateTeamInput struct {
	Name string `json:"name" binding:"required"`
}

// UpdateTeam memperbarui nama sebuah tim.
func UpdateTeam(c *gin.Context) {
	teamID := c.Param("teamId")

	var input UpdateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var team models.Team
	if err := config.DB.First(&team, teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	team.Name = input.Name
	config.DB.Save(&team)

	c.JSON(http.StatusOK, gin.H{"data": team})
}

// DeleteTeam menghapus sebuah tim beserta semua datanya.
func DeleteTeam(c *gin.Context) {
	teamID := c.Param("teamId")

	// Middleware sudah memastikan user adalah owner.
	// Kita akan menghapus tim dan GORM akan menangani relasi (cascade delete jika diatur).
	// Untuk amannya, kita bisa hapus todos terkait secara manual.
	tx := config.DB.Begin()

	// 1. Hapus todos di dalam tim
	if err := tx.Where("team_id = ?", teamID).Delete(&models.Todo{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todos in team"})
		return
	}

	// 2. Hapus asosiasi member (GORM biasanya menangani ini via many2many)
	// Kita bisa hapus manual untuk memastikan.
	if err := tx.Exec("DELETE FROM team_members WHERE team_id = ?", teamID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team members"})
		return
	}

	// 3. Hapus tim itu sendiri
	if err := tx.Where("id = ?", teamID).Delete(&models.Team{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}
