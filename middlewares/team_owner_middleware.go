// middlewares/team_owner_middleware.go
package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"notedteam.backend/config"
	"notedteam.backend/models"
)

func TeamOwnerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		teamID := c.Param("teamId")

		var team models.Team
		if err := config.DB.First(&team, teamID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}

		if team.OwnerID != userID.(uint) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only the team owner can perform this action"})
			return
		}

		c.Next()
	}
}
