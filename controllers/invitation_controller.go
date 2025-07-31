// controllers/invitation_controller.go
package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"notedteam.backend/config"
	"notedteam.backend/models"
)

// GetMyInvitations mengambil semua undangan yang tertunda untuk pengguna yang login.
func GetMyInvitations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var invitations []models.Invitation

	// Preload("Team") untuk menyertakan informasi tim dalam respons
	config.DB.Preload("Team").Where("user_id = ? AND status = ?", userID, models.InvitationPending).Find(&invitations)

	c.JSON(http.StatusOK, gin.H{"data": invitations})
}

// RespondToInvitation menangani accept/decline.
func RespondToInvitation(c *gin.Context) {
	userID, _ := c.Get("user_id")
	invitationID := c.Param("invitationId")

	var input struct {
		Accept bool `json:"accept"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. 'accept' must be true or false."})
		return
	}

	var invitation models.Invitation
	if err := config.DB.Where("id = ? AND user_id = ?", invitationID, userID).First(&invitation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found or you are not authorized"})
		return
	}

	if input.Accept {
		// --- TERIMA UNDANGAN ---
		// 1. Ambil objek user dari context dengan aman
		userContext, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User object not found in context"})
			return
		}

		// 2. Lakukan type assertion
		user, ok := userContext.(models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object type in context"})
			return
		}

		// 3. Jalankan Transaksi
		err := config.DB.Transaction(func(tx *gorm.DB) error {
			// Gunakan pointer ke objek user yang sudah di-assert (&user)
			if err := tx.Model(&models.Team{ID: invitation.TeamID}).Association("Members").Append(&user); err != nil {
				return err
			}

			invitation.Status = models.InvitationAccepted
			if err := tx.Save(&invitation).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			// Tambahkan log ini untuk debug di masa depan
			log.Printf("Failed to accept invitation: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process invitation acceptance"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted successfully"})

	} else {
		// --- TOLAK UNDANGAN ---
		invitation.Status = models.InvitationDeclined
		config.DB.Save(&invitation)
		c.JSON(http.StatusOK, gin.H{"message": "Invitation declined"})
	}
}
