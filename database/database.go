package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Open เปิด SQLite และสร้างตารางที่โปรแกรมต้องใช้
func Open() (*gorm.DB, error) {
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("สร้างโฟลเดอร์เก็บฐานข้อมูลไม่สำเร็จ: %w", err)
	}

	dbPath := filepath.Join("data", "photo-backup.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("เปิดฐานข้อมูลไม่สำเร็จ: %w", err)
	}

	if err := db.AutoMigrate(&Photo{}); err != nil {
		return nil, fmt.Errorf("สร้างตาราง photos ไม่สำเร็จ: %w", err)
	}

	return db, nil
}
