package main

import (
	"fmt"
	"photo-backup/database"
	"photo-backup/service"
)

// Resolve the ID against SQLite instead of exposing arbitrary local-file reads.
func (a *App) photoPath(photoID uint) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("ฐานข้อมูลยังไม่พร้อม")
	}
	var photo database.Photo
	if err := a.db.First(&photo, photoID).Error; err != nil {
		return "", fmt.Errorf("ไม่พบรูปที่เลือก: %w", err)
	}
	return photo.DestinationPath, nil
}

func (a *App) GetPhotoPreview(photoID uint) (string, error) {
	a.previewMu.Lock()
	defer a.previewMu.Unlock()
	path, err := a.photoPath(photoID)
	if err != nil {
		return "", err
	}
	return service.PhotoPreview(path)
}

// AnalyzePhoto returns draft tags; Nuxt explicitly saves them via SetPhotoTags.
func (a *App) AnalyzePhoto(photoID uint) ([]string, error) {
	if !a.analysisMu.TryLock() {
		return nil, fmt.Errorf("กำลังวิเคราะห์รูปอื่นอยู่ กรุณารอให้เสร็จ")
	}
	defer a.analysisMu.Unlock()
	path, err := a.photoPath(photoID)
	if err != nil {
		return nil, err
	}
	return service.AnalyzeImage(a.ctx, path)
}
