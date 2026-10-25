package database

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func normalizeTagName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// SetPhotoTags บันทึก Tag หลายคำให้รูปหนึ่งรูป
// ถ้า Tag มีอยู่แล้วจะนำมาใช้ซ้ำ และไม่สร้างความสัมพันธ์ซ้ำ
func SetPhotoTags(db *gorm.DB, photoID uint, tagNames []string) ([]Tag, error) {
	if photoID == 0 {
		return nil, errors.New("ต้องระบุรหัสรูป")
	}

	var photo Photo
	if err := db.First(&photo, photoID).Error; err != nil {
		return nil, err
	}

	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	seen := make(map[string]bool)
	tags := make([]Tag, 0, len(tagNames))

	for _, rawName := range tagNames {
		name := normalizeTagName(rawName)

		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		tag := Tag{Name: name}

		if err := tx.Where("name = ?", name).FirstOrCreate(&tag).Error; err != nil {
			return nil, err
		}

		relationship := PhotoTag{
			PhotoID: photoID,
			TagID:   tag.ID,
		}

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&relationship).Error; err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return tags, nil
}

// ListPhotoTags คืน Tag ทั้งหมดของรูปหนึ่งรูป
func ListPhotoTags(db *gorm.DB, photoID uint) ([]Tag, error) {
	var tags []Tag

	err := db.Table("tags").
		Select("tags.*").
		Joins("JOIN photo_tags ON photo_tags.tag_id = tags.id").
		Where("photo_tags.photo_id = ?", photoID).
		Order("tags.name ASC").
		Find(&tags).Error

	return tags, err
}

type PhotoTagItem struct {
	PhotoID uint   `json:"photoId"`
	ID      uint   `json:"id"`
	Name    string `json:"name"`
}

// GetAllPhotoTags คืน Tag ของทุกรูปในฐานข้อมูลเป็น map[photoID][]Tag
// ช่วยลดจำนวน query จาก N ครั้งเหลือครั้งเดียว
func GetAllPhotoTags(db *gorm.DB) (map[uint][]Tag, error) {
	var rows []PhotoTagItem

	err := db.Table("photo_tags").
		Select("photo_tags.photo_id, tags.id, tags.name").
		Joins("JOIN tags ON tags.id = photo_tags.tag_id").
		Order("tags.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint][]Tag)
	for _, r := range rows {
		result[r.PhotoID] = append(result[r.PhotoID], Tag{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	return result, nil
}

// SearchPhotosByTag ค้นหารูปจาก Tag หนึ่งคำ
func SearchPhotosByTag(db *gorm.DB, tagName string) ([]Photo, error) {
	name := normalizeTagName(tagName)
	if name == "" {
		return nil, errors.New("กรุณาระบุ Tag ที่ต้องการค้นหา")
	}

	var photos []Photo

	err := db.Model(&Photo{}).
		Distinct().
		Joins("JOIN photo_tags ON photo_tags.photo_id = photos.id").
		Joins("JOIN tags ON tags.id = photo_tags.tag_id").
		Where("tags.name = ?", name).
		Order("photos.backed_up_at DESC").
		Find(&photos).Error

	return photos, err
}

// GetTagCloud คืน Tag ที่ถูกใช้ พร้อมจำนวนรูปของแต่ละ Tag
func GetTagCloud(db *gorm.DB) ([]TagCloudItem, error) {
	var items []TagCloudItem

	err := db.Table("tags").
		Select("tags.id, tags.name, COUNT(photo_tags.photo_id) AS photo_count").
		Joins("JOIN photo_tags ON photo_tags.tag_id = tags.id").
		Group("tags.id, tags.name").
		Having("COUNT(photo_tags.photo_id) >= ?", 2).
		Order("photo_count DESC, tags.name ASC").
		Scan(&items).Error

	return items, err
}

// ClearPhotoTags ลบความสัมพันธ์ระหว่างรูปกับ Tag ทั้งหมดของรูปนี้
func ClearPhotoTags(db *gorm.DB, photoID uint) error {
	if photoID == 0 {
		return errors.New("ต้องระบุรหัสรูป")
	}

	return db.Where("photo_id = ?", photoID).Delete(&PhotoTag{}).Error
}
