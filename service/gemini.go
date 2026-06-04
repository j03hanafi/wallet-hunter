package service

import (
	"context"
	"fmt"
	"strings"
	"wallet-hunter/config"

	"google.golang.org/genai"
)

const (
	instruction = `
You are a strict OCR extractor. The image contains an Ethereum/EVM
private key (a 64-character hexadecimal string, optionally prefixed
with "0x").

Output ONLY the private key, exactly as it appears, on a single line.

Rules:
- No labels, no explanation, no markdown, no code fences, no quotes.
- Do not add or remove characters. Do not "correct" the key.
- If no 64-hex-character key is present, output exactly: NO_KEY_FOUND
`

	mime = "image/jpeg"
)

type Gemini struct {
	cfg    *config.Config
	Client *genai.Client
	Config *genai.GenerateContentConfig
}

func NewGemini(ctx context.Context, cfg *config.Config) (*Gemini, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	geminiCfg := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr[float32](0.0),
		TopP:            genai.Ptr[float32](0.0),
		TopK:            genai.Ptr[float32](1.0),
		CandidateCount:  1,
		MaxOutputTokens: 512,
		ThinkingConfig:  &genai.ThinkingConfig{ThinkingBudget: genai.Ptr[int32](0)},
	}

	return &Gemini{
		cfg:    cfg,
		Client: client,
		Config: geminiCfg,
	}, nil
}

func (g *Gemini) OCRImage(ctx context.Context, image []byte) (string, error) {

	parts := []*genai.Part{
		{InlineData: &genai.Blob{Data: image, MIMEType: mime}},
		{Text: instruction},
	}

	res, err := g.Client.Models.GenerateContent(ctx, g.cfg.GeminiModel, []*genai.Content{{Parts: parts}}, g.Config)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	text := strings.TrimSpace(res.Text())
	text = strings.TrimSpace(strings.Trim(text, "`\"'"))

	if text == "" || text == "NO_KEY_FOUND" {
		return "", fmt.Errorf("no valid text found in image")
	}

	return text, nil
}
