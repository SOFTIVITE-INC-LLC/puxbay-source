package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ImageService handles image processing and storage.
type ImageService struct {
	UploadDir string
}

func NewImageService() *ImageService {
	// Ensure upload directory exists
	dir := "uploads/images"
	os.MkdirAll(dir, os.ModePerm)
	return &ImageService{UploadDir: dir}
}

// OptimizeImage takes a raw image payload, resizes it if it exceeds max dimensions,
// and converts it to WebP format for optimal web delivery.
// Currently, this is a stub that simulates processing.
func (s *ImageService) OptimizeImage(data []byte, maxWidth, maxHeight int) ([]byte, string, error) {
	if len(data) == 0 {
		return nil, "", fmt.Errorf("empty image data")
	}

	// Simulate processing time
	// time.Sleep(100 * time.Millisecond)

	// In a real implementation, we would use a library like 'github.com/h2non/bimg'
	// or 'golang.org/x/image/webp' to resize and convert.
	// Example pseudo-code:
	// img := bimg.NewImage(data)
	// options := bimg.Options{Width: maxWidth, Height: maxHeight, Type: bimg.WEBP}
	// newImage, err := img.Process(options)

	// For the stub, we just return the original data and pretend it's webp
	// if it were really processed.

	// Create a dummy webp-like response for the stub if needed,
	// or just return the original bytes.

	var buf bytes.Buffer
	buf.Write(data)

	// In reality we would return the optimized bytes and "image/webp"
	return buf.Bytes(), "image/jpeg", nil
}

// UploadImageToStorage saves the image to the local disk.
func (s *ImageService) UploadImageToStorage(filename string, data io.Reader) (string, error) {
	filePath := filepath.Join(s.UploadDir, filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, data); err != nil {
		return "", err
	}

	return fmt.Sprintf("/%s", filePath), nil
}
