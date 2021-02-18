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
	Hash string
	TakenAt time.Time
}