package service

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseTags(t *testing.T) {
	got, err := parseTags(`{"tags":[" Cat ","cat","BEACH","","ท้องฟ้า"]}`)
	if err != nil || !reflect.DeepEqual(got, []string{"cat", "beach", "ท้องฟ้า"}) {
		t.Fatalf("%v %v", got, err)
	}
	for _, raw := range []string{`cat, sky`, `{"tags":[1]}`, `{}`, `{"tags":null}`} {
		if _, err := parseTags(raw); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	got, err = parseTags(`{"tags":[]}`)
	if err != nil || len(got) != 0 {
		t.Fatalf("%v %v", got, err)
	}
}

func TestAnalyzeProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/generate" {
			t.Errorf("bad request: %s %s", r.Method, r.URL.Path)
		}
		var request struct {
			Model  string   `json:"model"`
			Images []string `json:"images"`
			Stream bool     `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.Model != LocalModel || request.Stream || len(request.Images) != 1 || request.Images[0] != "aW1hZ2U=" {
			t.Errorf("%+v", request)
		}
		json.NewEncoder(w).Encode(map[string]any{"done": true, "response": `{"tags":[" Cat ","Sky"]}`})
	}))
	defer server.Close()
	got, err := analyzeWithClient(context.Background(), []byte("image"), server.Client(), server.URL)
	if err != nil || !reflect.DeepEqual(got, []string{"cat", "sky"}) {
		t.Fatalf("%v %v", got, err)
	}
}

func TestAnalyzeFailures(t *testing.T) {
	for _, item := range []struct {
		status int
		body   string
	}{
		{404, `{"error":"model missing"}`},
		{200, `{"done":false,"response":"{}"}`},
		{200, `{"done":true,"response":"not json"}`},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(item.status); w.Write([]byte(item.body)) }))
		_, err := analyzeWithClient(context.Background(), nil, server.Client(), server.URL)
		server.Close()
		if err == nil {
			t.Fatalf("accepted %+v", item)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := analyzeWithClient(ctx, nil, http.DefaultClient, "http://127.0.0.1:1"); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestImagePreparation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "photo.png")
	img := image.NewRGBA(image.Rect(0, 0, 120, 60))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	jpeg, err := ImageJPEG(path, 40)
	if err != nil {
		t.Fatal(err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(jpeg))
	if err != nil || format != "jpeg" || cfg.Width != 40 || cfg.Height != 20 {
		t.Fatalf("%+v %s %v", cfg, format, err)
	}
	if err := os.WriteFile(path, []byte("not an image"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImageJPEG(path, 40); err == nil {
		t.Fatal("accepted fake image")
	}
	if _, err := ImageJPEG(path+"missing", 40); err == nil {
		t.Fatal("accepted missing file")
	}
}
