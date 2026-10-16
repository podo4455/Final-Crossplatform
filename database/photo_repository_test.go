package database

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSaveAndListPhotos(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&Photo{}); err != nil {
		t.Fatal(err)
	}

	expected := &Photo{
		FileName:        "beach.jpg",
		SourcePath:      "E:/DCIM/beach.jpg",
		DestinationPath: "D:/Photos/beach.jpg",
		FileSize:        2048,
		BackupStatus:    "copied",
		BackedUpAt:      time.Date(2026, 10, 16, 10, 0, 0, 0, time.UTC),
	}

	if err := SavePhoto(db, expected); err != nil {
		t.Fatal(err)
	}

	photos, err := ListPhotos(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(photos) != 1 {
		t.Fatalf("ต้องมี 1 รูป แต่พบ %d", len(photos))
	}

	if photos[0].FileName != "beach.jpg" {
		t.Fatalf("ได้ชื่อไฟล์ %q", photos[0].FileName)
	}

	if photos[0].DestinationPath != "D:/Photos/beach.jpg" {
		t.Fatalf("ได้ตำแหน่งไฟล์ %q", photos[0].DestinationPath)
	}
}

func TestSetPhotoTagsAndSearch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&Photo{}, &Tag{}, &PhotoTag{}); err != nil {
		t.Fatal(err)
	}

	photo := &Photo{
		FileName:        "cat-beach.jpg",
		SourcePath:      "E:/DCIM/cat-beach.jpg",
		DestinationPath: "D:/Photos/cat-beach.jpg",
		FileSize:        2048,
		BackupStatus:    "copied",
		BackedUpAt:      time.Now().UTC(),
	}

	if err := SavePhoto(db, photo); err != nil {
		t.Fatal(err)
	}

	tags, err := SetPhotoTags(db, photo.ID, []string{"Cat", " cat ", "Beach"})
	if err != nil {
		t.Fatal(err)
	}

	if len(tags) != 2 {
		t.Fatalf("ต้องมี 2 Tag แต่พบ %d", len(tags))
	}

	found, err := SearchPhotosByTag(db, "CAT")
	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 1 || found[0].ID != photo.ID {
		t.Fatalf("ผลค้นหาไม่ตรง: %+v", found)
	}

	cloud, err := GetTagCloud(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(cloud) != 2 {
		t.Fatalf("ต้องมี 2 Tag ใน Tag Cloud แต่พบ %d", len(cloud))
	}
}
