package upload

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var AllowedImageTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

func ValidateImage(header *multipart.FileHeader, maxSizeMB int64) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !AllowedImageTypes[ext] {
		return fmt.Errorf("tipe file tidak didukung, hanya jpg/png yang diizinkan")
	}
	if header.Size > maxSizeMB*1024*1024 {
		return fmt.Errorf("ukuran file melebihi batas %dMB", maxSizeMB)
	}
	return nil
}

var AllowedVideoTypes = map[string]bool{
	".mp4": true,
	".mov": true,
}

func ValidateVideo(header *multipart.FileHeader, maxSizeMB int64) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !AllowedVideoTypes[ext] {
		return fmt.Errorf("tipe video tidak didukung, hanya mp4/mov yang diizinkan")
	}
	if header.Size > maxSizeMB*1024*1024 {
		return fmt.Errorf("ukuran video melebihi batas %dMB", maxSizeMB)
	}
	return nil
}

func GenerateFilename(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	return uuid.New().String() + ext
}
