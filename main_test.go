package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"photo-backup/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestParseEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := `
# ความคิดเห็น
GOOGLE_VISION_API_KEY=test-api-key-12345
ANOTHER_KEY="quoted-value"
SINGLE_QUOTED='single-value'
EMPTY_KEY=
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("เขียนไฟล์ทดสอบไม่สำเร็จ: %v", err)
	}

	// เคลียร์ค่าก่อนเทส
	os.Unsetenv("GOOGLE_VISION_API_KEY")
	os.Unsetenv("ANOTHER_KEY")
	os.Unsetenv("SINGLE_QUOTED")

	parseEnvFile(envPath)

	if val := os.Getenv("GOOGLE_VISION_API_KEY"); val != "test-api-key-12345" {
		t.Errorf("คาดหวัง test-api-key-12345 แต่ได้ %q", val)
	}
	if val := os.Getenv("ANOTHER_KEY"); val != "quoted-value" {
		t.Errorf("คาดหวัง quoted-value แต่ได้ %q", val)
	}
	if val := os.Getenv("SINGLE_QUOTED"); val != "single-value" {
		t.Errorf("คาดหวัง single-value แต่ได้ %q", val)
	}
}

func TestGetPhotoDataURLFollowsRenamedFile(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&database.Photo{}, &database.Tag{}, &database.PhotoTag{}); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	original := filepath.Join(dir, "original.jpg")
	if err := os.WriteFile(original, []byte("image-bytes"), 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := database.HashFile(original)
	if err != nil {
		t.Fatal(err)
	}

	photo := &database.Photo{
		FileName:        "original.jpg",
		SourcePath:      "E:/DCIM/original.jpg",
		DestinationPath: original,
		FileSize:        int64(len("image-bytes")),
		FileHash:        hash,
		BackupStatus:    "copied",
		BackedUpAt:      time.Now().UTC(),
	}
	if err := database.SavePhoto(db, photo); err != nil {
		t.Fatal(err)
	}

	renamed := filepath.Join(dir, "renamed.jpg")
	if err := os.Rename(original, renamed); err != nil {
		t.Fatal(err)
	}

	app := &App{db: db}
	dataURL, err := app.GetPhotoDataURL(photo.ID)
	if err != nil {
		t.Fatalf("อ่านรูปหลังเปลี่ยนชื่อไม่สำเร็จ: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/jpeg;base64,") {
		t.Fatalf("ได้ data URL ผิดรูปแบบ: %.40s", dataURL)
	}

	var saved database.Photo
	if err := db.First(&saved, photo.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.DestinationPath != renamed || saved.FileName != "renamed.jpg" {
		t.Fatalf("ฐานข้อมูลยังไม่ถูกอัปเดต: %+v", saved)
	}
}
