package main

import (
	"context"
	"fmt"
	"path/filepath"
	"photo-backup/service"
	"sync"
	"time"

	"photo-backup/database"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx      context.Context
	backupMu sync.Mutex
	db       *gorm.DB
}

// saveCopiedPhotos บันทึกเฉพาะไฟล์ที่ Copy สำเร็จลงฐานข้อมูล
func (a *App) saveCopiedPhotos(result service.Result) error {
	if a.db == nil {
		return fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	for _, file := range result.Files {
		if file.Status != "copied" {
			continue
		}

		copiedAt, err := time.Parse(time.RFC3339, file.CopiedAt)
		if err != nil {
			return fmt.Errorf("อ่านเวลาสำรองของ %s ไม่สำเร็จ: %w", file.Source, err)
		}

		photo := &database.Photo{
			FileName:        filepath.Base(file.Destination),
			SourcePath:      file.Source,
			DestinationPath: file.Destination,
			FileSize:        file.Size,
			BackupStatus:    "copied",
			BackedUpAt:      copiedAt,
		}

		if err := database.SavePhoto(a.db, photo); err != nil {
			return fmt.Errorf("บันทึกข้อมูลรูป %s ไม่สำเร็จ: %w", file.Source, err)
		}
	}

	return nil
}

// GetPhotos คืนรายการรูปที่เคยสำรองไว้ เพื่อส่งให้หน้า Gallery
func (a *App) GetPhotos() ([]database.Photo, error) {
	if a.db == nil {
		return nil, fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	return database.ListPhotos(a.db)
}

// BackupPhotos is the Wails application boundary; filesystem logic lives in service.
func (a *App) BackupPhotos(source, destination string) (service.Result, error) {
	if !a.backupMu.TryLock() {
		return service.Result{}, fmt.Errorf("กำลังสำรองรูปอยู่ กรุณารอให้เสร็จ")
	}
	defer a.backupMu.Unlock()

	result, err := service.CopyImages(a.ctx, source, destination, func(progress service.Progress) {
		runtime.EventsEmit(a.ctx, "backup:progress", progress)
	})
	if err != nil {
		return result, err
	}

	if err := a.saveCopiedPhotos(result); err != nil {
		return result, err
	}

	return result, nil
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	db, err := database.Open()
	if err != nil {
		runtime.LogError(ctx, err.Error())
		return
	}

	a.db = db
	runtime.LogInfo(ctx, "เปิดฐานข้อมูล SQLite สำเร็จ")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("สวัสดี %s — Nuxt เรียก Go ผ่าน Wails ได้แล้ว", name)
}

// SelectPhotos allows the user to select individual images together.
func (a *App) SelectPhotos() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "เลือกรูปภาพ (เลือกได้หลายรูป)",
		Filters: []runtime.FileFilter{{DisplayName: "รูปภาพ", Pattern: "*.jpg;*.jpeg;*.png;*.gif;*.webp;*.bmp;*.tif;*.tiff;*.heic;*.heif;*.avif"}},
	})
}

func (a *App) BackupPhoto(path, destination string) (service.Result, error) {
	return a.BackupSelectedPhotos([]string{path}, destination)
}

func (a *App) BackupSelectedPhotos(paths []string, destination string) (service.Result, error) {
	if !a.backupMu.TryLock() {
		return service.Result{}, fmt.Errorf("กำลังสำรองรูปอยู่ กรุณารอให้เสร็จ")
	}
	defer a.backupMu.Unlock()

	result, err := service.CopySelectedImages(a.ctx, paths, destination, func(progress service.Progress) {
		runtime.EventsEmit(a.ctx, "backup:progress", progress)
	})
	if err != nil {
		return result, err
	}

	if err := a.saveCopiedPhotos(result); err != nil {
		return result, err
	}

	return result, nil
}

func (a *App) SelectSource() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "เลือกโฟลเดอร์ต้นทาง",
	})
}

// SelectDestination เปิดหน้าต่างเลือกโฟลเดอร์ปลายทาง
func (a *App) SelectDestination() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "เลือกโฟลเดอร์ปลายทาง",
	})
}
