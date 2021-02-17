package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Conn *gorm.DB

func Init() {
	var err error
	Conn, err = gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	Conn.AutoMigrate(&Image{})
}