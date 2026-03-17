package parser

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"citadel/internal/recipe"
)

type scanIngredient struct {
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
	Item   string  `json:"item"`
}

type scanResult struct {
	Title           string           `json:"title"`
	Description     *string          `json:"description"`
	Ingredients     []scanIngredient `json:"ingredients"`
	Instructions    []string         `json:"instructions"`
	CookTimeMinutes *float64         `json:"cook_time_minutes"`
	Serves          *uint32          `json:"serves"`
	Cuisine         *string          `json:"cuisine"`
	Category        *string          `json:"category"`
}

func (s *scanResult) toRecipe() *recipe.Recipe {
	r := &recipe.Recipe{
		Title:        s.Title,
		Description:  s.Description,
		Instructions: s.Instructions,
	}

	for _, ing := range s.Ingredients {
		r.Ingredients = append(r.Ingredients, recipe.Ingredient{
			Amount: ing.Amount,
			Unit:   recipe.Unit(ing.Unit),
			Item:   ing.Item,
		})
	}

	if s.CookTimeMinutes != nil {
		d := time.Duration(*s.CookTimeMinutes * float64(time.Minute))
		r.CookTime = &d
	}

	r.Serves = s.Serves

	if s.Cuisine != nil {
		c := recipe.Cuisine(*s.Cuisine)
		if c.Valid() {
			r.Cuisine = &c
		}
	}

	if s.Category != nil {
		c := recipe.Category(*s.Category)
		if c.Valid() {
			r.Category = &c
		}
	}

	return r
}

//go:embed prompt.txt
var defaultPrompt string

type Claude struct {
	APIKey string
	Model  string
}

func New(apiKey, model string) *Claude {
	return &Claude{
		APIKey: apiKey,
		Model:  model,
	}
}

// Parse sends OCR text to Claude and returns a structured Recipe.
func (c *Claude) Parse(text string) (*recipe.Recipe, error) {
	reqBody := struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		System    string `json:"system"`
		Messages  []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}{
		Model:     c.Model,
		MaxTokens: 4096,
		System:    defaultPrompt,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: text},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		`https://api.anthropic.com/v1/messages`,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("API returned empty content")
	}

	// NOTE: even though we ask for JSON, the response may be wrapped in markdown code blocks, so we need to clean it up before parsing
	content := strings.TrimSpace(apiResp.Content[0].Text)
	if after, ok := strings.CutPrefix(content, "```json"); ok {
		content = strings.TrimSuffix(strings.TrimSpace(after), "```")
		content = strings.TrimSpace(content)
	} else if after, ok := strings.CutPrefix(content, "```"); ok {
		content = strings.TrimSuffix(strings.TrimSpace(after), "```")
		content = strings.TrimSpace(content)
	}

	var scan scanResult
	if err := json.Unmarshal([]byte(content), &scan); err != nil {
		return nil, fmt.Errorf(
			"failed to parse recipe JSON from Claude response: %w\nraw response: %s",
			err,
			content,
		)
	}

	return scan.toRecipe(), nil
}
