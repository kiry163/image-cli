package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kiry163/image-cli/pkg/apperror"
)

const (
	defaultVisionTimeout = 60 * time.Second
)

// VisionClient 视觉理解客户端
type VisionClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// VisionOptions 视觉分析选项
type VisionOptions struct {
	Model  string
	Prompt string
}

// VisionRequest 视觉分析请求
type VisionRequest struct {
	Model    string          `json:"model"`
	Messages []VisionMessage `json:"messages"`
}

// VisionMessage 视觉消息
type VisionMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// VisionContent 多模态内容项
type VisionContent struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL 图片URL
type ImageURL struct {
	URL string `json:"url"`
}

// VisionResponse 视觉分析响应
type VisionResponse struct {
	ID      string         `json:"id"`
	Created int            `json:"created"`
	Model   string         `json:"model"`
	Choices []VisionChoice `json:"choices"`
	Usage   VisionUsage    `json:"usage"`
}

// VisionChoice 视觉分析选项
type VisionChoice struct {
	Index        int           `json:"index"`
	Message      VisionMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// VisionUsage Token使用统计
type VisionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewVisionClient 创建视觉理解客户端
func NewVisionClient(apiKey, baseURL string) (*VisionClient, error) {
	if apiKey == "" {
		return nil, apperror.ConfigError("视觉理解 API Key 未配置", nil)
	}
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
	}

	return &VisionClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultVisionTimeout,
		},
	}, nil
}

// Analyze 分析图片
func (c *VisionClient) Analyze(ctx context.Context, imagePath string, opts VisionOptions) (string, error) {
	// 读取图片文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", apperror.InvalidInput("无法读取图片文件", err)
	}

	// base64编码
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	mimeType := detectMimeType(imagePath)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	// 构建请求
	reqBody := VisionRequest{
		Model: opts.Model,
		Messages: []VisionMessage{
			{
				Role: "user",
				Content: []VisionContent{
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL: imageURL,
						},
					},
					{
						Type: "text",
						Text: opts.Prompt,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", apperror.New("E401", "请求序列化失败", err.Error(), err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", apperror.New("E401", "创建请求失败", err.Error(), err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", apperror.New("E402", "API 请求失败", err.Error(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperror.New("E402", "读取响应失败", err.Error(), err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", apperror.New("E402", fmt.Sprintf("API 返回错误状态码: %d", resp.StatusCode), string(body), nil)
	}

	var visionResp VisionResponse
	if err := json.Unmarshal(body, &visionResp); err != nil {
		return "", apperror.New("E402", "解析响应失败", err.Error(), err)
	}

	if len(visionResp.Choices) == 0 {
		return "", apperror.New("E403", "API 未返回结果", "", nil)
	}

	// 获取文本内容
	if content, ok := visionResp.Choices[0].Message.Content.(string); ok {
		return content, nil
	}

	return "", apperror.New("E403", "无法获取分析结果", "", nil)
}
