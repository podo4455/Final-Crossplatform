package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"photo-backup/analysis"
	"photo-backup/database"

	"gorm.io/gorm"
)

const (
	defaultAnalysisThreshold = 0.75
	defaultAnalysisMaxTags   = 3
	defaultAnalysisWorkers   = 4
)

// AnalysisService ทำหน้าที่ประมวลผลวิเคราะห์รูปภาพ และแปลงผลเป็น Tag ในฐานข้อมูล
type AnalysisService struct {
	db        *gorm.DB
	analyzer  analysis.Analyzer
	Threshold float64
	MaxTags   int
	Workers   int
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
		Workers:   defaultAnalysisWorkers,
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

	// ถ้าไฟล์ถูกเปลี่ยนชื่อในโฟลเดอร์เดิม ให้เชื่อม path ใหม่ก่อนวิเคราะห์
	path, err := database.ResolvePhotoPath(s.db, &photo)
	if err != nil {
		return nil, err
	}

	// เรียก analyzer เพื่อดึง Label จากรูป (ยังไม่ลบ Tag เก่า จนกว่าจะได้ผลสำเร็จ)
	labels, err := s.analyzer.DetectLabels(ctx, path)
	if err != nil {
		// API ล้มเหลว — รักษา Tag เดิมไว้ ไม่ลบ
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
		// ไม่มี tag ผ่านเกณฑ์ — รักษา Tag เดิมไว้ คืน empty slice โดยไม่ error
		return []database.Tag{}, nil
	}

	// ล้าง Tag เก่าทิ้ง เฉพาะเมื่อได้ Tag ใหม่มาแทนแล้ว
	if err := database.ClearPhotoTags(s.db, photoID); err != nil {
		return nil, err
	}

	// บันทึก tag ใหม่ผ่าน repository ที่มีอยู่แล้ว
	tags, err := database.SetPhotoTags(s.db, photoID, tagNames)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// AnalyzePhotos วิเคราะห์หลายรูปพร้อมกันแบบ Worker Pool
func (s *AnalysisService) AnalyzePhotos(ctx context.Context, photoIDs []uint) (map[uint][]database.Tag, error) {
	if len(photoIDs) == 0 {
		return map[uint][]database.Tag{}, nil
	}

	workers := s.Workers
	if workers <= 0 {
		workers = defaultAnalysisWorkers
	}
	if workers > len(photoIDs) {
		workers = len(photoIDs)
	}

	result := make(map[uint][]database.Tag)
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, workers)

	for _, id := range photoIDs {
		if err := ctx.Err(); err != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(photoID uint) {
			defer wg.Done()
			defer func() { <-sem }()

			tags, err := s.AnalyzePhoto(ctx, photoID)
			if err != nil {
				return
			}
			if len(tags) == 0 {
				return
			}

			mu.Lock()
			result[photoID] = tags
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return result, nil
}
