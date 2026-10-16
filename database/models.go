package database

import "time"

// Photo คือข้อมูลเกี่ยวกับรูปที่สำรองแล้ว
// ไฟล์รูปจริงอยู่ในโฟลเดอร์ปลายทาง ไม่ได้เก็บลง SQLite
type Photo struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	FileName        string    `gorm:"not null" json:"fileName"`
	SourcePath      string    `gorm:"not null" json:"sourcePath"`
	DestinationPath string    `gorm:"not null;uniqueIndex" json:"destinationPath"`
	FileSize        int64     `gorm:"not null" json:"fileSize"`
	BackupStatus    string    `gorm:"not null" json:"backupStatus"`
	BackedUpAt      time.Time `json:"backedUpAt"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
