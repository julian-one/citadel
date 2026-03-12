package route

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"citadel/internal/ocr"
	"citadel/internal/parser"
)

func ScanRecipe(
	logger *slog.Logger,
	parser *parser.Claude,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Limit total request body to 20 MiB
		maxUploadSize := int64(20 << 20)
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		// Only buffer up to 2 MiB in RAM; larger files spill to disk
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				logger.Warn("upload exceeded maximum size", "limit", "20MB")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "file too large (max 20MB)"})
				return
			}
			logger.Error("failed to parse multipart form", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "failed to parse upload"})
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "no image file provided"})
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "unsupported file type: must be JPEG, PNG, or WEBP"})
			return
		}

		tmp, err := os.CreateTemp("", "recipe-scan-*"+ext)
		if err != nil {
			logger.Error("failed to create temp file", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}
		defer os.Remove(tmp.Name())
		defer tmp.Close()

		if _, err := io.Copy(tmp, file); err != nil {
			logger.Error("failed to write temp file", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}
		tmp.Close()

		logger.Info("running ocr", "file", header.Filename)
		text, err := ocr.Extract(ctx, tmp.Name())
		if err != nil {
			logger.Error("ocr failed", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "OCR processing failed: " + err.Error()})
			return
		}

		if strings.TrimSpace(text) == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "no text could be extracted from the image"})
			return
		}

		logger.Info("ocr complete", "text_length", len(text))

		logger.Info("parsing recipe with claude")
		recipe, err := parser.Parse(text)
		if err != nil {
			logger.Error("claude parsing failed", "error", err)
			if strings.Contains(err.Error(), "failed to parse recipe JSON") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "failed to parse recipe from extracted text"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "recipe parsing service unavailable"})
			return
		}

		logger.Info("recipe parsed", "title", recipe.Title)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipe)
	}
}
