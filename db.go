package main

import (
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Screenshot tabel for using to save screenshot data
type Screenshot struct {
	ID           int64 `gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	URL          string
	HTML         string `gorm:"type:text"`
	PlainText    string `gorm:"type:text"`
	ImgPath      string
	Phash        int64 `gorm:"index"`
	URLXxhash    int64 `gorm:"index"`
	TaskID       string
	TaskIDXxhash int64 `gorm:"index"`
}

// InitDb initialize database instance
func InitDb() (db *gorm.DB, err error) {
	dsn := viper.GetString("db.dsn")
	switch viper.GetString("db.dialect") {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		return
	}
	db.Debug()
	db.Debug().AutoMigrate(&Screenshot{})
	return
}
