package db

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	Source string
	DevicePath string
	Filename string
	Hash string
}