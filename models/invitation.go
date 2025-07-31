// models/invitation.go
package models

import "time"

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationDeclined InvitationStatus = "declined"
)

type Invitation struct {
	ID        uint             `json:"id" gorm:"primary_key"`
	UserID    uint             `json:"user_id"` // Siapa yang diundang
	TeamID    uint             `json:"team_id"` // Ke tim mana
	Status    InvitationStatus `json:"status" gorm:"type:enum('pending','accepted','declined');default:'pending'"`
	User      User             `json:"-" gorm:"foreignKey:UserID"`
	Team      Team             `json:"team" gorm:"foreignKey:TeamID"` // Sertakan data tim
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
