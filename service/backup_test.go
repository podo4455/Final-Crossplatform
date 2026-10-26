package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, path, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		t.Fatal(err)
	}
}
func TestCopyAndRepeat(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	put(t, filepath.Join(src, "เที่ยว", "PHOTO.JPG"), "original bytes")
	put(t, filepath.Join(src, "notes.txt"), "ignore")
	var last Progress
	result, err := CopyImages(context.Background(), src, dst, func(p Progress) { last = p })
	if err != nil || result.Copied != 1 || last.Completed != 1 {
		t.Fatalf("%+v %v %+v", result, err, last)
	}
	wantHash := sha256.Sum256([]byte("original bytes"))
	if result.Files[0].Hash != hex.EncodeToString(wantHash[:]) {
		t.Fatalf("hash ไม่ถูกต้อง: %q", result.Files[0].Hash)
	}
	for _, root := range []string{src, dst} {
		data, err := os.ReadFile(filepath.Join(root, "เที่ยว", "PHOTO.JPG"))
		if err != nil || string(data) != "original bytes" {
			t.Fatalf("%s: %q %v", root, data, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dst, "notes.txt")); !os.IsNotExist(err) {
		t.Fatal("copied non-image")
	}
	put(t, filepath.Join(src, "เที่ยว", "PHOTO.JPG"), "changed source")
	result, err = CopyImages(context.Background(), src, dst, nil)
	if err != nil || result.Skipped != 1 || result.Copied != 0 {
		t.Fatalf("%+v %v", result, err)
	}
	data, _ := os.ReadFile(filepath.Join(dst, "เที่ยว", "PHOTO.JPG"))
	if string(data) != "original bytes" {
		t.Fatal("overwrote destination")
	}
}

func TestSelectedImages(t *testing.T) {
	src, other, dst := t.TempDir(), t.TempDir(), t.TempDir()
	first := filepath.Join(src, "a.JPG")
	second := filepath.Join(other, "b.png")
	duplicate := filepath.Join(other, "a.JPG")
	text := filepath.Join(src, "notes.txt")
	for path, value := range map[string]string{first: "first", second: "second", duplicate: "duplicate", text: "text"} {
		put(t, path, value)
	}
	result, err := CopySelectedImages(context.Background(), []string{first, second, duplicate, text, src}, dst, nil)
	if err != nil || result.Copied != 2 || result.Skipped != 1 || result.Failed != 2 {
		t.Fatalf("%+v %v", result, err)
	}
	for path, want := range map[string]string{first: "first", second: "second", duplicate: "duplicate", filepath.Join(dst, "a.JPG"): "first", filepath.Join(dst, "b.png"): "second"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != want {
			t.Fatalf("%s: %q %v", path, data, err)
		}
	}
	result, err = CopySelectedImages(context.Background(), []string{first}, src, nil)
	if err != nil || result.Skipped != 1 {
		t.Fatalf("same destination: %+v %v", result, err)
	}
	if _, err := CopySelectedImages(context.Background(), nil, dst, nil); err == nil {
		t.Fatal("accepted empty selection")
	}
}
func TestRejectOverlappingFolders(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0755); err != nil {
		t.Fatal(err)
	}
	for _, paths := range [][2]string{{root, root}, {root, child}, {child, root}, {"", root}} {
		if _, err := CopyImages(context.Background(), paths[0], paths[1], nil); err == nil {
			t.Fatalf("accepted %v", paths)
		}
	}
}
func TestPartialFailureAndEmpty(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	result, err := CopyImages(context.Background(), src, dst, nil)
	if err != nil || result.Total != 0 || result.Files == nil {
		t.Fatalf("%+v %v", result, err)
	}
	put(t, filepath.Join(src, "blocked", "a.png"), "a")
	put(t, filepath.Join(src, "good.png"), "b")
	put(t, filepath.Join(dst, "blocked"), "existing file")
	result, err = CopyImages(context.Background(), src, dst, nil)
	if err != nil || result.Failed != 1 || result.Copied != 1 {
		t.Fatalf("%+v %v", result, err)
	}
}
func TestCancelledCopyRemovesPartialDestination(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	put(t, filepath.Join(src, "a.jpg"), "source remains")
	in, err := os.OpenRoot(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenRoot(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := copyFile(ctx, in, out, "a.jpg"); err == nil {
		t.Fatal("expected cancellation")
	}
	if _, err := os.Stat(filepath.Join(dst, "a.jpg")); !os.IsNotExist(err) {
		t.Fatal("partial file remains")
	}
	data, _ := os.ReadFile(filepath.Join(src, "a.jpg"))
	if string(data) != "source remains" {
		t.Fatal("source changed")
	}
}
func TestDestinationSymlinkCannotEscape(t *testing.T) {
	src, dst, outside := t.TempDir(), t.TempDir(), t.TempDir()
	put(t, filepath.Join(src, "linked", "a.jpg"), "photo")
	if err := os.Symlink(outside, filepath.Join(dst, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	result, err := CopyImages(context.Background(), src, dst, nil)
	if err != nil || result.Failed != 1 {
		t.Fatalf("%+v %v", result, err)
	}
	if _, err := os.Stat(filepath.Join(outside, "a.jpg")); !os.IsNotExist(err) {
		t.Fatal("escaped destination")
	}
}
