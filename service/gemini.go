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
You are a strict OCR extraction parser. Your task is to locate and extract an Ethereum/EVM private key from the provided image/text. 

A valid private key consists of exactly 64 hexadecimal characters (0-9, a-f, A-F), optionally preceded by "0x" (making it 66 characters). 

When evaluating the text to find the key, you MUST apply the following extraction rules internally:
1. Ignore Whitespace: Disregard all spaces, tabs, and line breaks. If a hex sequence spans multiple lines or has spaces in the middle, concatenate it.
2. Ignore Punctuation Noise: Disregard hyphens, dashes, periods, or commas that break up an otherwise valid hex string.
3. Ignore Surrounding Text: The key may be attached to labels (e.g., "Key:0x123..."). Extract ONLY the 64/66 character hex sequence.
4. Capitalization: Accept any mix of uppercase and lowercase hex characters.
5. Multiple Keys: If multiple valid keys are found, extract only the first one.

Output Rules (CRITICAL):
- Output ONLY the final, merged 64 or 66 character string on a single line.
- Absolutely NO labels, NO explanations, NO markdown formatting (no backticks), NO quotes, and NO code fences.
- Do not attempt to guess or correct non-hex characters (e.g., do not change an "O" to a "0"). 
- If no 64-character hexadecimal sequence can be assembled after applying the extraction rules above, output exactly: NO_KEY_FOUND
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
