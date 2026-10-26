package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// HashFile คำนวณ SHA-256 ของเนื้อไฟล์
// ใช้ตรวจว่ารูปที่ path เดิมหายไป ถูกเปลี่ยนชื่อเป็นไฟล์ไหนในโฟลเดอร์เดียวกัน
func HashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("อ่านไฟล์เพื่อคำนวณ hash ไม่สำเร็จ: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// isImageFileName ตรวจนามสกุลไฟล์รูปที่โปรแกรมรองรับ
func isImageFileName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tif", ".tiff", ".heic", ".heif", ".avif":
		return true
	}
	return false
}

// photoPathKey ใช้เทียบ path แบบไม่สนใจตัวพิมพ์เล็ก/ใหญ่ เพราะ Windows ไม่แยก
func photoPathKey(path string) string {
	return strings.ToLower(filepath.Clean(path))
}
