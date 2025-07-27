// models/todo.go
package models

import "time"

// Definisikan tipe kustom untuk status dan urgensi agar lebih terstruktur
type StatusType string
type UrgencyType string

const (
	StatusPending   StatusType = "pending"
	StatusWorking   StatusType = "working"
	StatusCompleted StatusType = "completed"
)

const (
	UrgencyLow    UrgencyType = "low"
	UrgencyMedium UrgencyType = "medium"
	UrgencyHigh   UrgencyType = "high"
)

type Todo struct {
	ID          uint        `json:"id" gorm:"primary_key"`
	Title       string      `json:"title" gorm:"not null"`
	Description string      `json:"description"`
	Status      StatusType  `json:"status" gorm:"type:enum('pending','working','completed');default:'pending'"`
	Urgency     UrgencyType `json:"urgency" gorm:"type:enum('low','medium','high');default:'low'"`
	UserID      uint        `json:"user_id"`                    // Foreign Key ke tabel User
	User        User        `json:"-" gorm:"foreignKey:UserID"` // Relasi ke User
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
