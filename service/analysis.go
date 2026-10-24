package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"photo-backup/analysis"
	"photo-backup/database"

	"gorm.io/gorm"
)

const (
	defaultAnalysisThreshold = 0.75
	defaultAnalysisMaxTags   = 3
)

// AnalysisService ทำหน้าที่ประมวลผลวิเคราะห์รูปภาพ และแปลงผลเป็น Tag ในฐานข้อมูล
type AnalysisService struct {
	db        *gorm.DB
	analyzer  analysis.Analyzer
	Threshold float64
	MaxTags   int
}

// NewAnalysisService สร้าง service สำหรับวิเคราะห์รูปภาพ
func NewAnalysisService(db *gorm.DB, analyzer analysis.Analyzer) *AnalysisService {
	if db == nil {
		panic("analysis service: db is nil")
	}
	if analyzer == nil {
		panic("analysis service: analyzer is nil")
	}

	return &AnalysisService{
		db:        db,
		analyzer:  analyzer,
		Threshold: defaultAnalysisThreshold,
		MaxTags:   defaultAnalysisMaxTags,
	}
}

// AnalyzePhoto วิเคราะห์รูปภาพ 1 รูป และบันทึก Tag ที่ผ่านเกณฑ์ลงฐานข้อมูล
func (s *AnalysisService) AnalyzePhoto(ctx context.Context, photoID uint) ([]database.Tag, error) {
	if photoID == 0 {
		return nil, errors.New("ต้องระบุรหัสรูป")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var photo database.Photo
	if err := s.db.First(&photo, photoID).Error; err != nil {
		return nil, err
	}

	if strings.TrimSpace(photo.DestinationPath) == "" {
		return nil, errors.New("ไม่พบตำแหน่งไฟล์รูปในฐานข้อมูล")
	}

	// ล้าง Tag เก่าทิ้งก่อน เพื่อให้ใช้ผลจากการวิเคราะห์ครั้งล่าสุดแทน
	if err := database.ClearPhotoTags(s.db, photoID); err != nil {
		return nil, err
	}

	// เรียก analyzer เพื่อดึง Label จากรูป
	labels, err := s.analyzer.DetectLabels(ctx, photo.DestinationPath)
	if err != nil {
		return nil, err
	}

	// เรียงตาม score มาก -> น้อย ก่อนกรอง เพื่อให้ได้แค่ Tag ที่มั่นใจที่สุด
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Score > labels[j].Score
	})

	// กรองตาม score และจำกัดจำนวน tag
	tagNames := make([]string, 0, s.MaxTags)
	seen := make(map[string]bool)

	for _, label := range labels {
		if s.Threshold > 0 && label.Score < s.Threshold {
			continue
		}

		name := strings.ToLower(strings.TrimSpace(label.Description))
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true

		tagNames = append(tagNames, name)
		if s.MaxTags > 0 && len(tagNames) >= s.MaxTags {
			break
		}
	}

	if len(tagNames) == 0 {
		// ถ้าไม่มี tag ผ่านเกณฑ์ ให้คืน empty slice โดยไม่ error
		return []database.Tag{}, nil
	}

	// บันทึก tag ผ่าน repository ที่มีอยู่แล้ว
	tags, err := database.SetPhotoTags(s.db, photoID, tagNames)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// AnalyzePhotos วิเคราะห์หลายรูปพร้อมกัน
func (s *AnalysisService) AnalyzePhotos(ctx context.Context, photoIDs []uint) (map[uint][]database.Tag, error) {
	result := make(map[uint][]database.Tag, len(photoIDs))

	for _, id := range photoIDs {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		tags, err := s.AnalyzePhoto(ctx, id)
		if err != nil {
			// ข้ามรูปที่ error แล้วไปต่อ ไม่ให้ทั้ง batch ล้มเหลว
			continue
		}
		if len(tags) > 0 {
			result[id] = tags
		}
	}

	return result, nil
}
