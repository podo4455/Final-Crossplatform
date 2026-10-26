package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
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

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Photo{}, &Tag{}, &PhotoTag{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func writePhoto(t *testing.T, path, content string) string {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSearchPhotosByMultipleTags(t *testing.T) {
	db := newTestDB(t)

	photos := make([]*Photo, 3)
	for i, name := range []string{"cat-beach.jpg", "cat-only.jpg", "beach-only.jpg"} {
		photo := &Photo{
			FileName:        name,
			SourcePath:      "E:/DCIM/" + name,
			DestinationPath: "D:/Photos/" + name,
			FileSize:        1024,
			BackupStatus:    "copied",
			BackedUpAt:      time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		if err := SavePhoto(db, photo); err != nil {
			t.Fatal(err)
		}
		photos[i] = photo
	}

	if _, err := SetPhotoTags(db, photos[0].ID, []string{"cat", "beach"}); err != nil {
		t.Fatal(err)
	}
	if _, err := SetPhotoTags(db, photos[1].ID, []string{"cat"}); err != nil {
		t.Fatal(err)
	}
	if _, err := SetPhotoTags(db, photos[2].ID, []string{"beach"}); err != nil {
		t.Fatal(err)
	}

	// OR: มี Tag ใดก็ได้ → ครบทั้ง 3 รูป
	any, err := SearchPhotosByTags(db, []string{"cat", "beach"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(any) != 3 {
		t.Fatalf("OR ต้องได้ 3 รูป แต่ได้ %d", len(any))
	}

	// AND: ต้องมีครบทุก Tag → เฉพาะรูปแรก
	all, err := SearchPhotosByTags(db, []string{"cat", "beach"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].ID != photos[0].ID {
		t.Fatalf("AND ต้องได้เฉพาะรูปแรก: %+v", all)
	}

	// ชื่อ Tag ซ้ำ/มีช่องว่าง/ตัวพิมพ์ต่างกัน ต้องถือเป็น Tag เดียว
	deduped, err := SearchPhotosByTags(db, []string{" Cat ", "CAT", "Beach"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(deduped) != 1 || deduped[0].ID != photos[0].ID {
		t.Fatalf("AND หลัง normalize ต้องได้เฉพาะรูปแรก: %+v", deduped)
	}

	// ไม่เลือก Tag → คืนรูปทั้งหมด
	empty, err := SearchPhotosByTags(db, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 3 {
		t.Fatalf("ไม่เลือก Tag ต้องได้ 3 รูป แต่ได้ %d", len(empty))
	}
}

func TestResolvePhotoPathFollowsRenamedFile(t *testing.T) {
	db := newTestDB(t)
	dir := t.TempDir()

	original := writePhoto(t, filepath.Join(dir, "old-name.jpg"), "photo-bytes")
	hash, err := HashFile(original)
	if err != nil {
		t.Fatal(err)
	}

	photo := &Photo{
		FileName:        "old-name.jpg",
		SourcePath:      "E:/DCIM/old-name.jpg",
		DestinationPath: original,
		FileSize:        int64(len("photo-bytes")),
		FileHash:        hash,
		BackupStatus:    "copied",
		BackedUpAt:      time.Now().UTC(),
	}
	if err := SavePhoto(db, photo); err != nil {
		t.Fatal(err)
	}

	renamed := filepath.Join(dir, "new-name.jpg")
	if err := os.Rename(original, renamed); err != nil {
		t.Fatal(err)
	}

	path, err := ResolvePhotoPath(db, photo)
	if err != nil {
		t.Fatal(err)
	}
	if path != renamed {
		t.Fatalf("ต้องได้ path ใหม่ %q แต่ได้ %q", renamed, path)
	}

	var saved Photo
	if err := db.First(&saved, photo.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.DestinationPath != renamed || saved.FileName != "new-name.jpg" {
		t.Fatalf("ฐานข้อมูลยังไม่ถูกอัปเดต: %+v", saved)
	}
}

func TestCleanupRelinksLegacyPhotoWithoutHash(t *testing.T) {
	db := newTestDB(t)
	dir := t.TempDir()

	original := writePhoto(t, filepath.Join(dir, "old.jpg"), "legacy-photo")
	photo := &Photo{
		FileName:        "old.jpg",
		SourcePath:      "E:/DCIM/old.jpg",
		DestinationPath: original,
		FileSize:        int64(len("legacy-photo")),
		BackupStatus:    "copied",
		BackedUpAt:      time.Now().UTC(),
	}
	if err := SavePhoto(db, photo); err != nil {
		t.Fatal(err)
	}

	renamed := filepath.Join(dir, "renamed.jpg")
	if err := os.Rename(original, renamed); err != nil {
		t.Fatal(err)
	}

	result, err := CleanupMissingPhotos(db)
	if err != nil {
		t.Fatal(err)
	}
	if result.Relinked != 1 || result.Removed != 0 {
		t.Fatalf("ต้องกู้คืน 1 รูปและไม่ลบ แต่ได้ %+v", result)
	}

	var saved Photo
	if err := db.First(&saved, photo.ID).Error; err != nil {
		t.Fatalf("record หายไป: %v", err)
	}
	if saved.DestinationPath != renamed {
		t.Fatalf("path ยังไม่ถูกอัปเดต: %q", saved.DestinationPath)
	}
	if saved.FileHash == "" {
		t.Fatal("ต้องเติม hash ให้ record เก่า")
	}
}

func TestCleanupRemovesPhotoWhoseFileIsGone(t *testing.T) {
	db := newTestDB(t)
	dir := t.TempDir()

	path := writePhoto(t, filepath.Join(dir, "gone.jpg"), "gone-photo")
	photo := &Photo{
		FileName:        "gone.jpg",
		SourcePath:      "E:/DCIM/gone.jpg",
		DestinationPath: path,
		FileSize:        int64(len("gone-photo")),
		BackupStatus:    "copied",
		BackedUpAt:      time.Now().UTC(),
	}
	if err := SavePhoto(db, photo); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	result, err := CleanupMissingPhotos(db)
	if err != nil {
		t.Fatal(err)
	}
	if result.Relinked != 0 || result.Removed != 1 {
		t.Fatalf("ต้องลบ 1 รายการ แต่ได้ %+v", result)
	}

	var count int64
	if err := db.Model(&Photo{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("ยังเหลือ record %d รายการ", count)
	}
}

func TestRelinkDoesNotStealFileOfAnotherPhoto(t *testing.T) {
	db := newTestDB(t)
	dir := t.TempDir()

	content := "same-size-content"
	first := writePhoto(t, filepath.Join(dir, "first.jpg"), content)
	second := writePhoto(t, filepath.Join(dir, "second.jpg"), content)

	for _, path := range []string{first, second} {
		hash, err := HashFile(path)
		if err != nil {
			t.Fatal(err)
		}
		photo := &Photo{
			FileName:        filepath.Base(path),
			SourcePath:      "E:/DCIM/" + filepath.Base(path),
			DestinationPath: path,
			FileSize:        int64(len(content)),
			FileHash:        hash,
			BackupStatus:    "copied",
			BackedUpAt:      time.Now().UTC(),
		}
		if err := SavePhoto(db, photo); err != nil {
			t.Fatal(err)
		}
	}

	// เปลี่ยนชื่อไฟล์แรก ไฟล์ที่สองยังอยู่ครบ
	renamed := filepath.Join(dir, "renamed.jpg")
	if err := os.Rename(first, renamed); err != nil {
		t.Fatal(err)
	}

	result, err := CleanupMissingPhotos(db)
	if err != nil {
		t.Fatal(err)
	}
	if result.Relinked != 1 || result.Removed != 0 {
		t.Fatalf("ต้องกู้คืน 1 รูปและไม่ลบ แต่ได้ %+v", result)
	}

	var savedSecond Photo
	if err := db.Where("file_name = ?", "second.jpg").First(&savedSecond).Error; err != nil {
		t.Fatal(err)
	}
	if savedSecond.DestinationPath != second {
		t.Fatalf("path ของรูปที่สองถูกแก้: %q", savedSecond.DestinationPath)
	}
}
