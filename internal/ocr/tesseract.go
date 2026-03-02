package ocr

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Extract runs tesseract on the given image file and returns the extracted text.
// It enforces a timeout of 10 seconds.
func Extract(ctx context.Context, imagePath string) (string, error) {
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
