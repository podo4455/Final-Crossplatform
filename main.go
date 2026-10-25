package main

import (
	"bufio"
	"embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// parseEnvFile อ่านไฟล์ .env และตั้งค่า environment variable ถ้ายังไม่มีการตั้งค่าไว้
func parseEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

// loadDotEnv ค้นหาและโหลดไฟล์ .env จากโฟลเดอร์ปัจจุบัน, โฟลเดอร์แม่ หรือตำแหน่ง executable
func loadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	// ค้นหาในโฟลเดอร์ปัจจุบันและไล่ขึ้นไป 3 ระดับ
	curr := dir
	for i := 0; i < 4; i++ {
		envPath := filepath.Join(curr, ".env")
		if info, err := os.Stat(envPath); err == nil && !info.IsDir() {
			parseEnvFile(envPath)
			return
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	// ค้นหาตำแหน่งเดียวกับ executable
	if exePath, err := os.Executable(); err == nil {
		exeDirEnv := filepath.Join(filepath.Dir(exePath), ".env")
		if info, err := os.Stat(exeDirEnv); err == nil && !info.IsDir() {
			parseEnvFile(exeDirEnv)
			return
		}
	}
}

func main() {
	loadDotEnv()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "photo-backup",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
