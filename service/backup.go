package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileResult struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Size        int64  `json:"size"`
	Status      string `json:"status"`
	Error       string `json:"error"`
	CopiedAt    string `json:"copiedAt"`
}
type Result struct {
	Total   int          `json:"total"`
	Copied  int          `json:"copied"`
	Skipped int          `json:"skipped"`
	Failed  int          `json:"failed"`
	Files   []FileResult `json:"files"`
}
type Progress struct {
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Current   string `json:"current"`
}

var errDestinationExists = errors.New("destination already exists")

func imageFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tif", ".tiff", ".heic", ".heif", ".avif":
		return true
	}
	return false
}

func directory(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("กรุณาเลือกโฟลเดอร์ต้นทางและปลายทาง")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve directory: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("ที่อยู่ที่เลือกต้องเป็นโฟลเดอร์")
	}
	return path, nil
}

// CopyImages keeps the relative folder structure and never overwrites existing files.
// Results are an integration boundary for the later database/tagging services.
func CopyImages(ctx context.Context, source, destination string, notify func(Progress)) (Result, error) {
	result := Result{Files: []FileResult{}}
	source, err := directory(source)
	if err != nil {
		return result, err
	}
	destination, err = directory(destination)
	if err != nil {
		return result, err
	}
	// Check both directions, including aliases on case-insensitive filesystems.
	for _, pair := range [][2]string{{source, destination}, {destination, source}} {
		ancestor, _ := os.Stat(pair[0])
		for path := pair[1]; ; path = filepath.Dir(path) {
			info, e := os.Stat(path)
			// A user may access a chosen folder without permission to stat its parents.
			if e == nil && os.SameFile(ancestor, info) {
				return result, errors.New("ต้นทางและปลายทางต้องเป็นคนละโฟลเดอร์ และไม่ซ้อนกัน")
			}
			if filepath.Dir(path) == path {
				break
			}
		}
	}
	src, err := os.OpenRoot(source)
	if err != nil {
		return result, err
	}
	defer src.Close()
	dst, err := os.OpenRoot(destination)
	if err != nil {
		return result, err
	}
	defer dst.Close()
	var names []string
	err = fs.WalkDir(src.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.Type().IsRegular() && imageFile(path) {
			names = append(names, filepath.FromSlash(path))
		}
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("อ่านต้นทางไม่สำเร็จ: %w", err)
	}
	result.Total = len(names)
	if notify != nil {
		notify(Progress{Total: result.Total})
	}
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		item := FileResult{Source: filepath.Join(source, name), Destination: filepath.Join(destination, name)}
		item.Size, err = copyFile(ctx, src, dst, name)
		switch {
		case errors.Is(err, errDestinationExists):
			item.Status = "skipped"
			result.Skipped++
		case err != nil:
			item.Status = "failed"
			item.Error = err.Error()
			result.Failed++
		default:
			item.Status = "copied"
			item.CopiedAt = time.Now().UTC().Format(time.RFC3339)
			result.Copied++
		}
		result.Files = append(result.Files, item)
		if notify != nil {
			notify(Progress{Total: result.Total, Completed: len(result.Files), Current: name})
		}
	}
	return result, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

// CopySelectedImages copies only the selected image files directly into destination.
func CopySelectedImages(ctx context.Context, paths []string, destination string, notify func(Progress)) (Result, error) {
	result := Result{Files: []FileResult{}}
	if len(paths) == 0 {
		return result, errors.New("กรุณาเลือกรูปภาพ")
	}
	destination, err := directory(destination)
	if err != nil {
		return result, err
	}
	dst, err := os.OpenRoot(destination)
	if err != nil {
		return result, err
	}
	defer dst.Close()
	result.Total = len(paths)
	if notify != nil {
		notify(Progress{Total: result.Total})
	}
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		item := FileResult{Source: path, Destination: filepath.Join(destination, filepath.Base(path))}
		item.Size, err = copySelectedFile(ctx, path, dst)
		switch {
		case errors.Is(err, errDestinationExists):
			item.Status = "skipped"
			result.Skipped++
		case err != nil:
			item.Status = "failed"
			item.Error = err.Error()
			result.Failed++
		default:
			item.Status = "copied"
			item.CopiedAt = time.Now().UTC().Format(time.RFC3339)
			result.Copied++
		}
		result.Files = append(result.Files, item)
		if notify != nil {
			notify(Progress{Total: result.Total, Completed: len(result.Files), Current: filepath.Base(path)})
		}
	}
	return result, nil
}

func copySelectedFile(ctx context.Context, path string, dst *os.Root) (int64, error) {
	if !filepath.IsAbs(path) || !imageFile(path) {
		return 0, errors.New("กรุณาเลือกไฟล์รูปภาพที่รองรับ")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	if !info.Mode().IsRegular() {
		return 0, errors.New("ต้องเป็นไฟล์รูปภาพปกติ ไม่ใช่โฟลเดอร์หรือ symbolic link")
	}
	src, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return 0, err
	}
	defer src.Close()
	return copyFile(ctx, src, dst, filepath.Base(path))
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func copyFile(ctx context.Context, src, dst *os.Root, name string) (size int64, err error) {
	in, err := src.Open(name)
	if err != nil {
		return 0, err
	}
	defer in.Close()
	before, err := in.Stat()
	if err != nil {
		return 0, err
	}
	if !before.Mode().IsRegular() {
		return 0, errors.New("ไม่ใช่ไฟล์ปกติ")
	}
	if err = dst.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return 0, err
	}
	out, err := dst.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if errors.Is(err, fs.ErrExist) {
		info, statErr := dst.Lstat(name)
		if statErr == nil && info.Mode().IsRegular() {
			return 0, errDestinationExists
		}
	}
	if err != nil {
		return 0, err
	}
	// Only remove the newly created destination if copying fails. Never delete a source.
	defer func() {
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			if removeErr := dst.Remove(name); removeErr != nil {
				err = fmt.Errorf("%w; ลบไฟล์ปลายทางที่ไม่สมบูรณ์ไม่ได้: %v", err, removeErr)
			}
		}
	}()
	size, err = io.Copy(out, contextReader{ctx, in})
	if err != nil {
		return size, err
	}
	after, err := in.Stat()
	if err != nil {
		return size, err
	}
	if size != before.Size() || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return size, errors.New("ไฟล์ต้นทางเปลี่ยนระหว่างคัดลอก กรุณาลองใหม่")
	}
	err = out.Sync()
	return size, err
}
