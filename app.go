package main

import (
	"context"
	"fmt"
	"photo-backup/service"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	backupMu sync.Mutex
}

// BackupPhotos is the Wails application boundary; filesystem logic lives in service.
func (a *App) BackupPhotos(source, destination string) (service.Result, error) {
	if !a.backupMu.TryLock() {
		return service.Result{}, fmt.Errorf("กำลังสำรองรูปอยู่ กรุณารอให้เสร็จ")
	}
	defer a.backupMu.Unlock()
	return service.CopyImages(a.ctx, source, destination, func(progress service.Progress) {
		runtime.EventsEmit(a.ctx, "backup:progress", progress)
	})
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
	return service.CopySelectedImages(a.ctx, paths, destination, func(progress service.Progress) {
		runtime.EventsEmit(a.ctx, "backup:progress", progress)
	})
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
