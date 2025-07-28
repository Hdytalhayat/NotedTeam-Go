// models/team.go
package models

import "time"

type Team struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	Name      string    `json:"name" gorm:"not null"`
	OwnerID   uint      `json:"owner_id"` // User yang membuat tim
	Owner     User      `json:"-" gorm:"foreignKey:OwnerID"`
	Members   []User    `json:"members,omitempty" gorm:"many2many:team_members;"`
	Todos     []Todo    `json:"todos,omitempty"` // Sebuah tim memiliki banyak todo
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
