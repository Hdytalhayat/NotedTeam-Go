// ws/hub.go
package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message mendefinisikan struktur pesan yang akan di-broadcast
type Message struct {
	Event string      `json:"event"` // e.g., "todo_created", "todo_updated"
	Data  interface{} `json:"data"`
}

// Client adalah representasi dari satu koneksi websocket
type Client struct {
	Conn   *websocket.Conn
	Send   chan []byte
	TeamID uint
}

// Hub mengelola semua client dan broadcast pesan
type Hub struct {
	Clients    map[uint]map[*Client]bool // Peta dari TeamID ke peta Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan struct { // Struct untuk membawa pesan dan TeamID
		Message []byte
		TeamID  uint
	}
	mu sync.Mutex // Untuk melindungi akses ke map Clients
}

// Global instance dari Hub
var AppHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast: make(chan struct {
			Message []byte
			TeamID  uint
		}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.Clients[client.TeamID]; !ok {
				h.Clients[client.TeamID] = make(map[*Client]bool)
			}
			h.Clients[client.TeamID][client] = true
			log.Printf("Client registered to team %d. Total clients in team: %d", client.TeamID, len(h.Clients[client.TeamID]))
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.TeamID]; ok {
				if _, ok := h.Clients[client.TeamID][client]; ok {
					delete(h.Clients[client.TeamID], client)
					close(client.Send)
					if len(h.Clients[client.TeamID]) == 0 {
						delete(h.Clients, client.TeamID)
					}
					log.Printf("Client unregistered from team %d. Total clients in team: %d", client.TeamID, len(h.Clients[client.TeamID]))
				}
			}
			h.mu.Unlock()

		case broadcast := <-h.Broadcast:
			h.mu.Lock()
			if clients, ok := h.Clients[broadcast.TeamID]; ok {
				for client := range clients {
					select {
					case client.Send <- broadcast.Message:
					default:
						// Gagal mengirim, mungkin koneksi terputus. Unregister client.
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}
