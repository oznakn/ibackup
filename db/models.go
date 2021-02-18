package db

import (
	"gorm.io/gorm"
	"time"
)

type Image struct {
	gorm.Model
	Source string
	DevicePath string
	Filename string `gorm:"index"`
	Hash string `gorm:"index"`
	Size uint
	TakenAt time.Time
}