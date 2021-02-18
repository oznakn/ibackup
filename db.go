package main

import (
	"gorm.io/driver/sqlite"
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

var Conn *gorm.DB

func dbInit(path string) {
	var err error
	Conn, err = gorm.Open(sqlite.Open(path), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	Conn.AutoMigrate(&Image{})
}