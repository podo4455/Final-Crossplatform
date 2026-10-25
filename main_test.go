package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := `
# ความคิดเห็น
GOOGLE_VISION_API_KEY=test-api-key-12345
ANOTHER_KEY="quoted-value"
SINGLE_QUOTED='single-value'
EMPTY_KEY=
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("เขียนไฟล์ทดสอบไม่สำเร็จ: %v", err)
	}

	// เคลียร์ค่าก่อนเทส
	os.Unsetenv("GOOGLE_VISION_API_KEY")
	os.Unsetenv("ANOTHER_KEY")
	os.Unsetenv("SINGLE_QUOTED")

	parseEnvFile(envPath)

	if val := os.Getenv("GOOGLE_VISION_API_KEY"); val != "test-api-key-12345" {
		t.Errorf("คาดหวัง test-api-key-12345 แต่ได้ %q", val)
	}
	if val := os.Getenv("ANOTHER_KEY"); val != "quoted-value" {
		t.Errorf("คาดหวัง quoted-value แต่ได้ %q", val)
	}
	if val := os.Getenv("SINGLE_QUOTED"); val != "single-value" {
		t.Errorf("คาดหวัง single-value แต่ได้ %q", val)
	}
}
