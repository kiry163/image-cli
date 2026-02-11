package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/sashabaranov/go-openai"
)

type OCRClient struct {
	client *openai.Client
	model  string
}

type OCROptions struct {
	Mode string
}

func NewOCRClient(apiKey, baseURL, model string) (*OCRClient, error) {
	if apiKey == "" {
		return nil, apperror.ConfigError("OCR API Key 未配置", nil)
	}
	if baseURL == "" {
		baseURL = "https://www.dmxapi.cn/v1"
	}
	if model == "" {
		model = "DeepSeek-OCR"
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &OCRClient{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}, nil
}

func (c *OCRClient) Recognize(ctx context.Context, imagePath string, opts OCROptions) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", apperror.InvalidInput("无法读取图片文件", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	mimeType := detectMimeType(imagePath)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	systemPrompt := buildOCRPrompt(opts.Mode)

	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role: "user",
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: imageURL,
						},
					},
				},
			},
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", apperror.New("E201", "OCR API 调用失败", err.Error(), err)
	}

	if len(resp.Choices) == 0 {
		return "", apperror.New("E201", "OCR API 返回空结果", "", nil)
	}

	return resp.Choices[0].Message.Content, nil
}

func buildOCRPrompt(mode string) string {
	switch strings.ToLower(mode) {
	case "markdown":
		return "<image>\n<|grounding|>Convert the document to markdown."
	case "text":
		return "<image>\n<|grounding|>OCR this image."
	case "figure":
		return "<image>\nParse the figure."
	case "detail":
		return "<image>\nDescribe this image in detail."
	case "free":
		fallthrough
	default:
		return "<image>\nFree OCR."
	}
}

func detectMimeType(path string) string {
	ext := strings.ToLower(path)
	switch {
	case strings.HasSuffix(ext, ".jpg"), strings.HasSuffix(ext, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(ext, ".png"):
		return "image/png"
	case strings.HasSuffix(ext, ".webp"):
		return "image/webp"
	case strings.HasSuffix(ext, ".gif"):
		return "image/gif"
	default:
		return "image/jpeg"
	}
}
