package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"photo-backup/analysis"
	"photo-backup/service"
	"sync"
	"time"

	"photo-backup/database"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx       context.Context
	backupMu  sync.Mutex
	db        *gorm.DB
	analysis  *service.AnalysisService
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

// SetPhotoTags บันทึก Tag ที่ได้จากระบบวิเคราะห์ภาพ
func (a *App) SetPhotoTags(photoID uint, tagNames []string) ([]database.Tag, error) {
	if a.db == nil {
		return nil, fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	return database.SetPhotoTags(a.db, photoID, tagNames)
}

// GetPhotoTags คืน Tag ของรูปหนึ่งรูป
func (a *App) GetPhotoTags(photoID uint) ([]database.Tag, error) {
	if a.db == nil {
		return nil, fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	return database.ListPhotoTags(a.db, photoID)
}

// SearchPhotosByTag ค้นหารูปที่มี Tag ตรงกับคำค้นหา
func (a *App) SearchPhotosByTag(tagName string) ([]database.Photo, error) {
	if a.db == nil {
		return nil, fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	return database.SearchPhotosByTag(a.db, tagName)
}

// GetTagCloud คืน Tag ทั้งหมดพร้อมจำนวนรูปที่เกี่ยวข้อง
func (a *App) GetTagCloud() ([]database.TagCloudItem, error) {
	if a.db == nil {
		return nil, fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	return database.GetTagCloud(a.db)
}

// AnalyzePhoto วิเคราะห์รูปภาพ 1 รูป และบันทึก Tag ลงฐานข้อมูล
func (a *App) AnalyzePhoto(photoID uint) ([]database.Tag, error) {
	if a.analysis == nil {
		return nil, fmt.Errorf("ระบบวิเคราะห์ภาพยังไม่พร้อม")
	}

	return a.analysis.AnalyzePhoto(a.ctx, photoID)
}

// AnalyzePhotos วิเคราะห์หลายรูปพร้อมกัน
func (a *App) AnalyzePhotos(photoIDs []uint) (map[uint][]database.Tag, error) {
	if a.analysis == nil {
		return nil, fmt.Errorf("ระบบวิเคราะห์ภาพยังไม่พร้อม")
	}

	return a.analysis.AnalyzePhotos(a.ctx, photoIDs)
}

// SetAnalysisAPIKey เปลี่ยน analyzer เป็น VisionAnalyzer เมื่อมี API key
func (a *App) SetAnalysisAPIKey(apiKey string) error {
	if a.db == nil {
		return fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}
	if apiKey == "" {
		a.analysis = service.NewAnalysisService(a.db, analysis.MockAnalyzer{})
		runtime.LogInfo(a.ctx, "เปลี่ยนเป็น MockAnalyzer")
		return nil
	}

	key := apiKey
	if key == "" {
		key = os.Getenv("GOOGLE_VISION_API_KEY")
	}
	if key == "" {
		return fmt.Errorf("กรุณาระบุ API key")
	}

	analyzer := analysis.VisionAnalyzer{APIKey: key}
	a.analysis = service.NewAnalysisService(a.db, analyzer)
	runtime.LogInfo(a.ctx, "เปลี่ยนเป็น Google Vision Analyzer")
	return nil
}

// GetPhotoDataURL คืนรูปในรูปแบบ data URL เพื่อแสดงใน Gallery
func (a *App) GetPhotoDataURL(photoID uint) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}

	var photo database.Photo
	if err := a.db.First(&photo, photoID).Error; err != nil {
		return "", err
	}

	if photo.DestinationPath == "" {
		return "", fmt.Errorf("ไม่พบตำแหน่งไฟล์รูป")
	}

	data, err := os.ReadFile(photo.DestinationPath)
	if err != nil {
		return "", fmt.Errorf("อ่านไฟล์รูปไม่สำเร็�: %w", err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(photo.DestinationPath))
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:" + mimeType + ";base64," + encoded, nil
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

	// ถ้ามี GOOGLE_VISION_API_KEY ให้ใช้ Google Cloud Vision ไม่งั้นใช้ MockAnalyzer
	var analyzer analysis.Analyzer
	if key := os.Getenv("GOOGLE_VISION_API_KEY"); key != "" {
		analyzer = analysis.VisionAnalyzer{APIKey: key}
		runtime.LogInfo(ctx, "เริ่มต้นระบบวิเคราะห์ภาพ (Google Cloud Vision API)")
	} else {
		analyzer = analysis.MockAnalyzer{}
		runtime.LogInfo(ctx, "เริ่มต้นระบบวิเคราะห์ภาพ (MockAnalyzer - ยังไม่มี GOOGLE_VISION_API_KEY)")
	}

	a.analysis = service.NewAnalysisService(db, analyzer)
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
