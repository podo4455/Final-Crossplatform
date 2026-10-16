package database

import (
	"errors"
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

// ListPhotos คืนรายการรูปที่บันทึกไว้ โดยเรียงรายการใหม่สุดก่อน
func ListPhotos(db *gorm.DB) ([]Photo, error) {
	var photos []Photo

	err := db.Order("backed_up_at DESC").Find(&photos).Error
	if err != nil {
		return nil, err
	}

	return photos, nil
}
