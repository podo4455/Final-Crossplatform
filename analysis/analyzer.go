package analysis

import (
	"context"
	"fmt"
)

// Label คือคำอธิบายหนึ่งคำที่ได้จากการวิเคราะห์ภาพ
type Label struct {
	Description string
	Score       float64
}

// Analyzer คือสัญญาว่า "อะไรก็ได้ที่อ่านไฟล์รูปแล้วคืน Label ได้"
// service จะคุยกับตัวนี้เท่านั้น ไม่รู้จักผู้ให้บริการข้างนอก
type Analyzer interface {
	DetectLabels(ctx context.Context, imagePath string) ([]Label, error)
}

// MockAnalyzer ใช้ระหว่างพัฒนา/สาธิตตอนยังไม่มี API key
type MockAnalyzer struct{}

func (MockAnalyzer) DetectLabels(ctx context.Context, imagePath string) ([]Label, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if imagePath == "" {
		return nil, fmt.Errorf("ต้องระบุตำแหน่งไฟล์รูป")
	}

	return []Label{
		{Description: "Cat", Score: 0.96},
		{Description: "Sofa", Score: 0.88},
		{Description: "Furniture", Score: 0.42},
	}, nil
}
