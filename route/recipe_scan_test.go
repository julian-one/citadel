package route

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newScanHandler() http.HandlerFunc {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Pass nil parser — we only test error paths that don't reach the parser.
	return ScanRecipe(logger, nil)
}

func postMultipart(
	t *testing.T,
	handler http.HandlerFunc,
	filename string,
	content []byte,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if filename != "" {
		part, err := writer.CreateFormFile("image", filename)
		require.NoError(t, err)
		_, err = part.Write(content)
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/recipes/scan", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestScanRecipe_NoFile(t *testing.T) {
	handler := newScanHandler()

	// Send multipart with no file field
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/recipes/scan", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "no image file provided", resp["error"])
}

func TestScanRecipe_UnsupportedFileType(t *testing.T) {
	handler := newScanHandler()
	rr := postMultipart(t, handler, "recipe.gif", []byte("fake gif"))

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "unsupported file type")
}

func TestScanRecipe_UnsupportedBMP(t *testing.T) {
	handler := newScanHandler()
	rr := postMultipart(t, handler, "recipe.bmp", []byte("fake bmp"))

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "unsupported file type")
}

func TestScanRecipe_AcceptedExtensions(t *testing.T) {
	// These extensions should pass the file type check.
	// They will fail later (OCR on invalid content), but that's fine —
	// we're testing that the extension filter doesn't reject them.
	exts := []string{".jpg", ".jpeg", ".png", ".webp"}

	handler := newScanHandler()

	for _, ext := range exts {
		t.Run(ext, func(t *testing.T) {
			rr := postMultipart(
				t,
				handler,
				"recipe"+ext,
				[]byte("not a real image"),
			)
			// Should NOT be 400 "unsupported file type"
			assert.NotEqual(t, http.StatusBadRequest, rr.Code,
				"extension %s should be accepted", ext)
		})
	}
}

func TestScanRecipe_ResponseIsJSON(t *testing.T) {
	handler := newScanHandler()
	rr := postMultipart(t, handler, "recipe.gif", []byte("fake"))

	assert.Equal(
		t,
		"application/json",
		rr.Header().Get("Content-Type"),
		"error responses should be JSON",
	)
}
