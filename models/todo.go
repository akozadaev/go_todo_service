package models

import "gorm.io/gorm"

type Todo struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

func (t *Todo) BeforeCreate(tx *gorm.DB) error {
	return nil
}
