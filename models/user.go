// models/user.go
package models

import "time"

type User struct {
	ID                    uint       `json:"id" gorm:"primary_key"`
	Name                  string     `json:"name" gorm:"not null"`
	Email                 string     `json:"email" gorm:"unique;not null"`
	Password              string     `json:"-" gorm:"not null"` // Tanda - agar tidak tampil di JSON
	IsVerified            bool       `json:"is_verified" gorm:"default:false"`
	VerificationToken     string     `json:"-" gorm:"size:255"`
	PasswordResetToken    string     `json:"-" gorm:"size:255"`
	PasswordResetTokenExp *time.Time `json:"-"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	Teams                 []Team     `json:"teams,omitempty" gorm:"many2many:team_members;"`
}
