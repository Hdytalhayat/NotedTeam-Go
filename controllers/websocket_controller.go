// controllers/websocket_controller.go
package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"notedteam.backend/config"
	"notedteam.backend/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Izinkan koneksi dari semua origin (sesuaikan untuk produksi!)
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ServeWs menangani permintaan koneksi websocket.
// Rute: GET /ws/teams/:teamId?token=...
func ServeWs(c *gin.Context) {
	// 1. Otentikasi & Otorisasi
	// Kita tetap butuh userID, tapi sekarang kita bisa yakin itu ada.
	userID, _ := c.Get("user_id")

	teamIdStr := c.Param("teamId")
	teamId, err := strconv.ParseUint(teamIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Team ID"})
		return
	}

	// Cek apakah user adalah member dari tim ini (logika dari TeamMemberMiddleware)
	var memberCount int64
	if err := config.DB.Table("team_members").Where("user_id = ? AND team_id = ?", userID, uint(teamId)).Count(&memberCount).Error; err != nil || memberCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this team"})
		return
	}

	// 2. Upgrade koneksi HTTP ke WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	// 3. Buat objek Client dan daftarkan ke Hub
	client := &ws.Client{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		TeamID: uint(teamId),
	}
	ws.AppHub.Register <- client

	// 4. Jalankan goroutine untuk membaca dan menulis pesan
	go writePump(client)
	go readPump(client)
}

func readPump(client *ws.Client) {
	defer func() {
		ws.AppHub.Unregister <- client
		client.Conn.Close()
	}()
	for {
		// Baca pesan dari koneksi (kita tidak mengharapkan pesan dari client untuk saat ini)
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func writePump(client *ws.Client) {
	defer func() {
		client.Conn.Close()
	}()
	for {
		message, ok := <-client.Send
		if !ok {
			// Hub menutup channel ini
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		// Kirim pesan ke client
		client.Conn.WriteMessage(websocket.TextMessage, message)
	}
}
