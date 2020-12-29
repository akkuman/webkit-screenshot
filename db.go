package main

import (
	"time"

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
func InitDb(dbpath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		return db, err
	}
	db.Debug()
	db.Debug().AutoMigrate(&Screenshot{})
	return db, nil
}
