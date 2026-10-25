package database

import (
	"errors"
	"os"
	"strings"

	"gorm.io/gorm"
)

// SavePhoto บันทึกข้อมูลของรูปที่สำรองสำเร็จลงตาราง photos
func SavePhoto(db *gorm.DB, photo *Photo) error {
	if strings.TrimSpace(photo.FileName) == "" {
		return errors.New("ต้องระบุชื่อไฟล์")
	}

	if strings.TrimSpace(photo.DestinationPath) == "" {
		return errors.New("ต้องระบุตำแหน่งปลายทาง")
	}

	return db.Create(photo).Error
}

// DeletePhoto ลบข้อมูลรูปและ Tag ที่เชื่อมอยู่ออกจากฐานข้อมูล
func DeletePhoto(db *gorm.DB, photoID uint) error {
	if photoID == 0 {
		return errors.New("ต้องระบุรหัสรูป")
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	// ลบความสัมพันธ์ Tag ก่อน
	if err := tx.Where("photo_id = ?", photoID).Delete(&PhotoTag{}).Error; err != nil {
		return err
	}

	// ลบ record รูป
	if err := tx.Delete(&Photo{}, photoID).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

// CleanupMissingPhotos สแกนรูปทั้งหมดในฐานข้อมูล
// ถ้าไฟล์จริงไม่อยู่แล้ว จะลบ record และ Tag ที่เชื่อมอยู่ออก
// คืนรายการ ID ที่ถูกลบ
func CleanupMissingPhotos(db *gorm.DB) ([]uint, error) {
	var photos []Photo
	if err := db.Find(&photos).Error; err != nil {
		return nil, err
	}

	var removed []uint

	for _, photo := range photos {
		if _, err := os.Stat(photo.DestinationPath); err == nil {
			continue // ไฟล์ยังอยู่
		}

		// ไฟล์ไม่อยู่แล้ว — ลบ record
		if err := DeletePhoto(db, photo.ID); err != nil {
			return removed, err
		}
		removed = append(removed, photo.ID)
	}

	return removed, nil
}

// ListPhotos คืนรายการรูปที่บันทึกไว้ โดยเรียงรายการใหม่สุดก่อน
func ListPhotos(db *gorm.DB) ([]Photo, error) {
	var photos []Photo

	err := db.Order("backed_up_at DESC").Find(&photos).Error
	if err != nil {
		return nil, err
	}

	return photos, nil
}
