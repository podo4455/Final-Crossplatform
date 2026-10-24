package analysis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestVisionAnalyzerDetectLabelsWithMockServer(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(imgPath, []byte("fake-jpeg-data"), 0644); err != nil {
		t.Fatalf("สร้างไฟล์ทดสอบไม่สำเร็จ: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"responses": [
				{
					"labelAnnotations": [
						{"description": "Cat", "score": 0.96, "topicality": 0.96},
						{"description": "Sofa", "score": 0.88, "topicality": 0.88},
						{"description": "Furniture", "score": 0.42, "topicality": 0.42}
					]
				}
			]
		}`))
	}))
	defer server.Close()

	analyzer := VisionAnalyzer{
		Client:   server.Client(),
		Endpoint: server.URL + "/v1/images:annotate?key=test-key",
	}

	labels, err := analyzer.DetectLabels(context.Background(), imgPath)
	if err != nil {
		t.Fatalf("ไม่คาดว่าจะ error: %v", err)
	}

	if len(labels) != 3 {
		t.Fatalf("คาดหวัง 3 labels แต่ได้ %d", len(labels))
	}

	if labels[0].Description != "Cat" {
		t.Fatalf("คาดหวัง Cat แต่ได้ %s", labels[0].Description)
	}
	if labels[0].Score != 0.96 {
		t.Fatalf("คาดหวัง score 0.96 แต่ได้ %f", labels[0].Score)
	}

	if labels[2].Description != "Furniture" || labels[2].Score != 0.42 {
		t.Fatalf("label สุดท้ายไม่ตรงตามที่คาดหวัง: %+v", labels[2])
	}

	t.Logf("ได้ labels สำเร็จ: %+v", labels)
}

func TestVisionAnalyzerEmptyImagePath(t *testing.T) {
	analyzer := VisionAnalyzer{}
	_, err := analyzer.DetectLabels(context.Background(), "")
	if err == nil {
		t.Fatal("ควรคืน error เมื่อ imagePath ว่าง")
	}
}
