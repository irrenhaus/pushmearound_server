package models

import (
	"github.com/jinzhu/gorm"
)

type AccessToken struct {
	gorm.Model
	UserID uint   `gorm:"index"`
	Token  string `gorm:"unique_index"`
}
