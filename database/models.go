package database

import "time"

// Photo คือข้อมูลเกี่ยวกับรูปที่สำรองแล้ว
// ไฟล์รูปจริงอยู่ในโฟลเดอร์ปลายทาง ไม่ได้เก็บลง SQLite
type Photo struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	FileName        string `gorm:"not null" json:"fileName"`
	SourcePath      string `gorm:"not null" json:"sourcePath"`
	DestinationPath string `gorm:"not null;uniqueIndex" json:"destinationPath"`
	FileSize        int64  `gorm:"not null" json:"fileSize"`
	// FileHash คือ SHA-256 ของเนื้อไฟล์ ใช้ตามหารูปที่ถูกเปลี่ยนชื่อในโฟลเดอร์เดิม
	// record เก่าอาจยังว่างอยู่ ระบบจะเติมให้เมื่ออ่านไฟล์สำเร็จ
	FileHash     string    `gorm:"index" json:"fileHash"`
	BackupStatus string    `gorm:"not null" json:"backupStatus"`
	BackedUpAt   time.Time `json:"backedUpAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CleanupResult สรุปผลการตรวจไฟล์ที่หายไป
type CleanupResult struct {
	// Relinked คือจำนวนรูปที่ถูกเปลี่ยนชื่อ แล้วระบบเชื่อมกลับให้อัตโนมัติ
	Relinked int `json:"relinked"`
	// Removed คือจำนวน record ที่ไฟล์จริงหายไปและลบออกจากฐานข้อมูลแล้ว
	Removed int `json:"removed"`
}

// Tag คือคำที่ใช้อธิบายหรือค้นหารูป เช่น แมว, ทะเล, อาหาร
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// PhotoTag เชื่อมรูปหนึ่งรูปกับ Tag หลายคำ
// รูปหนึ่งรูปมีหลาย Tag และ Tag หนึ่งคำใช้กับหลายรูปได้
type PhotoTag struct {
	PhotoID   uint      `gorm:"primaryKey" json:"photoId"`
	TagID     uint      `gorm:"primaryKey" json:"tagId"`
	CreatedAt time.Time `json:"createdAt"`
}

// TagCloudItem คือข้อมูลหนึ่งคำสำหรับแสดง Tag Cloud
type TagCloudItem struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	PhotoCount int64  `json:"photoCount"`
}
