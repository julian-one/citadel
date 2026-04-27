package ocr

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

// Extract runs tesseract on the given image file and returns the extracted text.
// It normalises EXIF orientation before OCR so that rotated phone photos are
// read correctly, then enforces a timeout of 10 seconds for the tesseract call.
func Extract(ctx context.Context, imagePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(imagePath))
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" {
		if err := normalizeOrientation(imagePath); err != nil {
			return "", fmt.Errorf("failed to normalize image orientation: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tesseract", imagePath, "stdout")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("tesseract timed out")
		}
		return "", fmt.Errorf("tesseract failed: %w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// normalizeOrientation opens the image with EXIF auto-orientation enabled
// and writes the pixel-corrected result back to the same path. This ensures
// tesseract (which ignores EXIF orientation tags) sees an upright image.
func normalizeOrientation(imagePath string) error {
	img, err := imaging.Open(imagePath, imaging.AutoOrientation(true))
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	if err := imaging.Save(img, imagePath); err != nil {
		return fmt.Errorf("failed to save corrected image: %w", err)
	}

	return nil
}
