package ocr_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"citadel/internal/ocr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtract_CrabCakes(t *testing.T) {
	// This test requires the crab-cakes.jpeg fixture which has EXIF
	// orientation 6 (90° CW). Without orientation correction the OCR
	// output is garbled.
	const fixture = "../../../crab-cakes.jpeg"

	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("fixture not found: crab-cakes.jpeg")
	}

	// Copy the fixture to a temp file so normalizeOrientation doesn't
	// mutate the original.
	data, err := os.ReadFile(fixture)
	require.NoError(t, err)

	tmp, err := os.CreateTemp("", "ocr-test-*.jpeg")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())

	_, err = tmp.Write(data)
	require.NoError(t, err)
	tmp.Close()

	text, err := ocr.Extract(context.Background(), tmp.Name())
	require.NoError(t, err)

	upper := strings.ToUpper(text)

	// Title
	assert.Contains(t, upper, "CRAB CAKES", "should extract the full title")
	assert.Contains(t, upper, "BENEDICT", "should extract the recipe title")
	assert.Contains(t, upper, "AVOCADO", "should extract 'with Avocado'")

	// All three components must appear
	assert.Contains(t, upper, "HOLLANDAISE", "should extract the Hollandaise component")
	assert.Contains(t, upper, "CRAB", "should extract the Crab Cakes component")
	assert.Contains(t, upper, "EGGS", "should extract the Eggs component")

	// Key ingredients from each component
	assert.Contains(t, upper, "SRIRACHA", "should extract Sriracha ingredient")
	assert.Contains(t, upper, "MAYONNAISE", "should extract mayonnaise ingredient")
	assert.Contains(t, upper, "AVOCADO", "should extract avocado ingredient")

	// Serves info
	assert.Contains(t, upper, "SERVES 4", "should extract serving info")

	// Instructions
	assert.Contains(t, upper, "SKILLET", "should extract cooking instructions")

	// Time metadata
	assert.Contains(t, upper, "PREP TIME", "should extract prep time label")
	assert.Contains(t, upper, "TOTAL TIME", "should extract total time label")
	assert.Contains(t, upper, "30 MINUTES", "should extract prep time value")

	// Source (page footer)
	assert.Contains(t, upper, "CRAVINGS", "should extract cookbook source")

	// Component grouping labels
	assert.Contains(t, upper, "FOR THE", "should detect component grouping markers")

	// Instruction section headers
	assert.Contains(
		t,
		upper,
		"MAKE THE FAKE HOLLANDAISE",
		"should detect hollandaise instruction header",
	)
	assert.Contains(t, upper, "MAKE THE CRAB CAKES", "should detect crab cakes instruction header")
	assert.Contains(t, upper, "POACH THE EGGS", "should detect eggs instruction header")

	t.Logf("OCR text length: %d characters", len(text))
	t.Logf("First 500 chars:\n%s", text[:min(500, len(text))])
}
