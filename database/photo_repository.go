package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// SavePhoto บันทึกข้อมูลของรูปที่สำรองสำเร็จลงตาราง photos
// ถ้ามี destination_path ซ้ำอยู่แล้ว จะข้ามโดยไม่ error (กันซ้ำ)
func SavePhoto(db *gorm.DB, photo *Photo) error {
	if strings.TrimSpace(photo.FileName) == "" {
		return errors.New("ต้องระบุชื่อไฟล์")
	}

	if strings.TrimSpace(photo.DestinationPath) == "" {
		return errors.New("ต้องระบุตำแหน่งปลายทาง")
	}

	var existing Photo
	err := db.Where("destination_path = ?", photo.DestinationPath).First(&existing).Error
	if err == nil {
		// เคยบันทึกไฟล์นี้ไปแล้ว ไม่ต้องบันทึกซ้ำ
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
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
// ถ้าไฟล์ถูกเปลี่ยนชื่อในโฟลเดอร์เดิม จะเชื่อม record กลับให้อัตโนมัติ
// ถ้าไฟล์หายไปจริง จะลบ record และ Tag ที่เชื่อมอยู่ออก
func CleanupMissingPhotos(db *gorm.DB) (CleanupResult, error) {
	var result CleanupResult

	var photos []Photo
	if err := db.Find(&photos).Error; err != nil {
		return result, err
	}

	var missing []Photo
	for _, photo := range photos {
		if strings.TrimSpace(photo.DestinationPath) == "" {
			missing = append(missing, photo)
			continue
		}

		info, err := os.Stat(photo.DestinationPath)
		if err == nil && info.Mode().IsRegular() {
			continue // ไฟล์ยังอยู่
		}

		missing = append(missing, photo)
	}

	relinked, err := relinkMissingPhotos(db, missing)
	if err != nil {
		return result, err
	}
	result.Relinked = len(relinked)

	for _, photo := range missing {
		if _, ok := relinked[photo.ID]; ok {
			continue
		}

		if err := DeletePhoto(db, photo.ID); err != nil {
			return result, err
		}
		result.Removed++
	}

	return result, nil
}

// RelinkMissingPhotos ตรวจรูปที่ไฟล์หาย แล้วพยายามเชื่อมกลับกับไฟล์ที่ถูกเปลี่ยนชื่อ
// ในโฟลเดอร์เดิม คืนจำนวนรูปที่เชื่อมกลับสำเร็จ
func RelinkMissingPhotos(db *gorm.DB) (int, error) {
	var photos []Photo
	if err := db.Select("id", "file_name", "destination_path", "file_size", "file_hash").Find(&photos).Error; err != nil {
		return 0, err
	}

	var missing []Photo
	for _, photo := range photos {
		if strings.TrimSpace(photo.DestinationPath) == "" {
			continue
		}
		if info, err := os.Stat(photo.DestinationPath); err == nil && info.Mode().IsRegular() {
			continue
		}
		missing = append(missing, photo)
	}

	if len(missing) == 0 {
		return 0, nil
	}

	relinked, err := relinkMissingPhotos(db, missing)
	if err != nil {
		return 0, err
	}

	return len(relinked), nil
}

// ResolvePhotoPath คืน path ของไฟล์รูปที่อ่านได้จริง
// ถ้าไฟล์เดิมถูกเปลี่ยนชื่อในโฟลเดอร์เดียวกัน จะเชื่อม record กลับให้อัตโนมัติ
// พร้อมเติม FileHash ให้ record เก่าที่เคยบันทึกไว้ก่อนมีระบบ hash
func ResolvePhotoPath(db *gorm.DB, photo *Photo) (string, error) {
	if photo == nil {
		return "", errors.New("ไม่พบข้อมูลรูป")
	}
	if strings.TrimSpace(photo.DestinationPath) == "" {
		return "", errors.New("ไม่พบตำแหน่งไฟล์รูป")
	}

	if info, err := os.Stat(photo.DestinationPath); err == nil && info.Mode().IsRegular() {
		if photo.FileHash == "" {
			if hash, hashErr := HashFile(photo.DestinationPath); hashErr == nil {
				if updateErr := db.Model(&Photo{}).Where("id = ?", photo.ID).Update("file_hash", hash).Error; updateErr == nil {
					photo.FileHash = hash
				}
			}
		}
		return photo.DestinationPath, nil
	}

	relinked, err := relinkMissingPhotos(db, []Photo{*photo})
	if err != nil {
		return "", err
	}

	if path, ok := relinked[photo.ID]; ok {
		photo.DestinationPath = path
		photo.FileName = filepath.Base(path)
		return path, nil
	}

	return "", fmt.Errorf("ไม่พบไฟล์รูป %s ในโฟลเดอร์ %s", photo.FileName, filepath.Dir(photo.DestinationPath))
}

// relinkMissingPhotos พยายามจับคู่ record ที่ไฟล์หายไปกับไฟล์ที่ถูกเปลี่ยนชื่อ
// โดยดูจากขนาดไฟล์และ SHA-256 เฉพาะไฟล์ในโฟลเดอร์เดิมที่ยังไม่มี record ใดอ้างถึง
// คืน map[photoID]newPath ของรูปที่เชื่อมกลับสำเร็จ
func relinkMissingPhotos(db *gorm.DB, photos []Photo) (map[uint]string, error) {
	relinked := make(map[uint]string)
	if len(photos) == 0 {
		return relinked, nil
	}

	// โหลด path ที่มีอยู่ในฐานข้อมูลครั้งเดียว เพื่อไม่ให้แย่งไฟล์ที่ record อื่นใช้อยู่
	var existing []Photo
	if err := db.Select("id", "destination_path").Find(&existing).Error; err != nil {
		return nil, err
	}
	referenced := make(map[string]bool, len(existing))
	for _, item := range existing {
		if item.DestinationPath != "" {
			referenced[photoPathKey(item.DestinationPath)] = true
		}
	}

	// จัดกลุ่มรูปที่ไฟล์หายตามโฟลเดอร์ เพื่อสแกนแต่ละโฟลเดอร์ครั้งเดียว
	groups := make(map[string][]Photo)
	for _, photo := range photos {
		if photo.DestinationPath == "" {
			continue
		}
		dir := filepath.Dir(photo.DestinationPath)
		groups[dir] = append(groups[dir], photo)
	}

	for dir, group := range groups {
		candidates := unregisteredImages(dir, referenced)
		if len(candidates) == 0 {
			continue
		}

		bySize := make(map[int64][]string)
		for _, path := range candidates {
			info, err := os.Stat(path)
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			bySize[info.Size()] = append(bySize[info.Size()], path)
		}

		hashes := make(map[string]string)
		hashOf := func(path string) (string, bool) {
			if hash, ok := hashes[path]; ok {
				return hash, hash != ""
			}
			hash, err := HashFile(path)
			if err != nil {
				hashes[path] = ""
				return "", false
			}
			hashes[path] = hash
			return hash, true
		}

		for _, photo := range group {
			pool := bySize[photo.FileSize]
			if photo.FileSize <= 0 {
				for _, paths := range bySize {
					pool = append(pool, paths...)
				}
			}
			if len(pool) == 0 {
				continue
			}

			var target string
			if photo.FileHash != "" {
				// ตรวจ hash ให้แน่ใจว่าเป็นไฟล์เดียวกันจริง
				for _, path := range pool {
					if hash, ok := hashOf(path); ok && hash == photo.FileHash {
						target = path
						break
					}
				}
			} else if len(pool) == 1 {
				// record เก่าที่ยังไม่มี hash: จับคู่ด้วยขนาดไฟล์เมื่อมีผู้สมัครเพียงไฟล์เดียว
				target = pool[0]
			}

			if target == "" {
				continue
			}

			updates := map[string]any{
				"destination_path": target,
				"file_name":        filepath.Base(target),
			}
			if photo.FileHash == "" {
				if hash, ok := hashOf(target); ok {
					updates["file_hash"] = hash
				}
			}

			if err := db.Model(&Photo{}).Where("id = ?", photo.ID).Updates(updates).Error; err != nil {
				continue
			}

			referenced[photoPathKey(target)] = true
			relinked[photo.ID] = target
			if photo.FileSize > 0 {
				bySize[photo.FileSize] = removePath(bySize[photo.FileSize], target)
			} else {
				for size, paths := range bySize {
					bySize[size] = removePath(paths, target)
				}
			}
		}
	}

	return relinked, nil
}

// unregisteredImages คืนไฟล์รูปในโฟลเดอร์ที่ยังไม่มี record ใดในฐานข้อมูลอ้างถึง
func unregisteredImages(dir string, referenced map[string]bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !isImageFileName(entry.Name()) {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if referenced[photoPathKey(path)] {
			continue
		}
		files = append(files, path)
	}

	return files
}

// removePath คืน slice ใหม่ที่ตัด path ที่ระบุออก
func removePath(paths []string, target string) []string {
	result := paths[:0]
	removed := false
	for _, path := range paths {
		if !removed && photoPathKey(path) == photoPathKey(target) {
			removed = true
			continue
		}
		result = append(result, path)
	}
	return result
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

// SearchPhotosByName ค้นหารูปจากชื่อไฟล์
func SearchPhotosByName(db *gorm.DB, keyword string) ([]Photo, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return ListPhotos(db)
	}

	var photos []Photo
	err := db.Where("LOWER(file_name) LIKE LOWER(?)", "%"+keyword+"%").
		Order("backed_up_at DESC").
		Find(&photos).Error
	if err != nil {
		return nil, err
	}

	return photos, nil
}
