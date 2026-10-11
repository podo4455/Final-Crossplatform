package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("สวัสดี %s — Nuxt เรียก Go ผ่าน Wails ได้แล้ว", name)
}

// SelectSource เปิดหน้าต่างเลือกโฟลเดอร์ต้นทาง
func (a *App) SelectSource() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "เลือกโฟลเดอร์ต้นทาง",
	})
}

// SelectDestination เปิดหน้าต่างเลือกโฟลเดอร์ปลายทาง
func (a *App) SelectDestination() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "เลือกโฟลเดอร์ปลายทาง",
	})
}
