package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"photo-backup/database"
)

func TestPhotoLookupAndTagWorkflow(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "photos.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.AutoMigrate(&database.Photo{}, &database.Tag{}, &database.PhotoTag{}); err != nil {
		t.Fatal(err)
	}
	app := NewApp()
	app.db = db
	app.ctx = context.Background()
	if _, err := app.GetPhotoPreview(999); err == nil {
		t.Fatal("unknown photo ID accepted")
	}
	photo := &database.Photo{FileName: "cat.jpg", DestinationPath: filepath.Join(t.TempDir(), "cat.jpg"), BackupStatus: "copied", BackedUpAt: time.Now()}
	if err := database.SavePhoto(db, photo); err != nil {
		t.Fatal(err)
	}
	if _, err := app.AnalyzePhoto(photo.ID); err == nil {
		t.Fatal("missing image accepted")
	}
	if _, err := app.SetPhotoTags(photo.ID, []string{" Cat ", "cat", "SKY"}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.SetPhotoTags(photo.ID, []string{"sky", "beach"}); err != nil {
		t.Fatal(err)
	}
	tags, err := app.GetPhotoTags(photo.ID)
	if err != nil || len(tags) != 3 {
		t.Fatalf("%+v %v", tags, err)
	}
	photos, err := app.SearchPhotosByTag(" CAT ")
	if err != nil || len(photos) != 1 || photos[0].ID != photo.ID {
		t.Fatalf("%+v %v", photos, err)
	}
	cloud, err := app.GetTagCloud()
	if err != nil || len(cloud) != 3 {
		t.Fatalf("%+v %v", cloud, err)
	}
	for _, item := range cloud {
		if item.PhotoCount != 1 {
			t.Fatalf("duplicate relationship: %+v", item)
		}
	}
}

// Opt-in real-model smoke test. It uses a temporary database, never the user's library.
func TestLocalVisionIntegration(t *testing.T) {
	path := os.Getenv("PHOTO_BACKUP_TEST_IMAGE")
	if path == "" {
		t.Skip("set PHOTO_BACKUP_TEST_IMAGE to run the real local model")
	}
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "vision.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.AutoMigrate(&database.Photo{}, &database.Tag{}, &database.PhotoTag{}); err != nil {
		t.Fatal(err)
	}
	photo := &database.Photo{FileName: filepath.Base(path), DestinationPath: path, BackupStatus: "copied", BackedUpAt: time.Now()}
	if err := database.SavePhoto(db, photo); err != nil {
		t.Fatal(err)
	}
	app := NewApp()
	app.db = db
	app.ctx = context.Background()
	photos, err := app.GetPhotos()
	if err != nil || len(photos) != 1 {
		t.Fatalf("%v %v", photos, err)
	}
	if _, err := app.GetPhotoPreview(photos[0].ID); err != nil {
		t.Fatal(err)
	}
	tags, err := app.AnalyzePhoto(photos[0].ID)
	if err != nil || len(tags) == 0 {
		t.Fatalf("%v %v", tags, err)
	}
	t.Logf("Real local model tags: %v", tags)
	if _, err := app.SetPhotoTags(photo.ID, tags); err != nil {
		t.Fatal(err)
	}
	stored, err := app.GetPhotoTags(photo.ID)
	if err != nil || len(stored) != len(tags) {
		t.Fatalf("%v %v", stored, err)
	}
	found, err := app.SearchPhotosByTag(tags[0])
	if err != nil || len(found) != 1 || found[0].ID != photo.ID {
		t.Fatalf("%v %v", found, err)
	}
	cloud, err := app.GetTagCloud()
	if err != nil || len(cloud) != len(tags) {
		t.Fatalf("%v %v", cloud, err)
	}
}
