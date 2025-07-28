// middlewares/team_member_middleware.go
package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"notedteam.backend/config"
)

func TeamMemberMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		teamID := c.Param("teamId")

		var memberCount int64
		// Hitung apakah ada relasi antara userID dan teamID di tabel team_members
		err := config.DB.Table("team_members").
			Where("user_id = ? AND team_id = ?", userID, teamID).
			Count(&memberCount).Error

		if err != nil || memberCount == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not a member of this team"})
			return
		}

		c.Next()
	}
}
