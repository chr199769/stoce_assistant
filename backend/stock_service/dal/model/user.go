package model

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;type:varchar(64)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
