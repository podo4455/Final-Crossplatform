package analysis

import (
	"context"
	"testing"
)

func TestMockAnalyzerReturnsLabels(t *testing.T) {
	labels, err := MockAnalyzer{}.DetectLabels(context.Background(), "test.jpg")
	if err != nil {
		t.Fatalf("ไม่คาดว่าจะ error: %v", err)
	}

	if len(labels) == 0 {
		t.Fatal("Mock ต้องคืน Label อย่างน้อยหนึ่งคำ")
	}

	t.Logf("ได้ %d labels: %+v", len(labels), labels)
}
