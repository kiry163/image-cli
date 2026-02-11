package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kiry163/image-cli/pkg/apperror"
)

const (
	defaultGenerateTimeout = 60 * time.Second
)

// ImageGenerationClient 图像生成客户端
type ImageGenerationClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// GenerateOptions 图像生成选项
type GenerateOptions struct {
	Model   string // cogview-3-flash, glm-image, cogview-4, cogview-4-250304
	Size    string // 1024x1024, 768x1344, etc.
	Quality string // standard, hd
}

// GenerateRequest 图像生成请求
type GenerateRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Size    string `json:"size,omitempty"`
	Quality string `json:"quality,omitempty"`
}

// GenerateResponse 图像生成响应
type GenerateResponse struct {
	Created       int             `json:"created"`
	Data          []ImageData     `json:"data"`
	ContentFilter []ContentFilter `json:"content_filter,omitempty"`
}

// ImageData 生成的图片数据
type ImageData struct {
	URL string `json:"url"`
}

// ContentFilter 内容过滤信息
type ContentFilter struct {
	Role  string `json:"role"`
	Level int    `json:"level"`
}

// NewImageGenerationClient 创建图像生成客户端
func NewImageGenerationClient(apiKey, baseURL string) (*ImageGenerationClient, error) {
	if apiKey == "" {
		return nil, apperror.ConfigError("图像生成 API Key 未配置", nil)
	}
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
	}

	return &ImageGenerationClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultGenerateTimeout,
		},
	}, nil
}

// Generate 根据提示词生成图像
func (c *ImageGenerationClient) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	if prompt == "" {
		return "", apperror.InvalidArgument("提示词不能为空", nil)
	}

	reqBody := GenerateRequest{
		Model:   opts.Model,
		Prompt:  prompt,
		Size:    opts.Size,
		Quality: opts.Quality,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", apperror.New("E301", "请求序列化失败", err.Error(), err)
	}

	url := fmt.Sprintf("%s/images/generations", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", apperror.New("E301", "创建请求失败", err.Error(), err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", apperror.New("E302", "API 请求失败", err.Error(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperror.New("E302", "读取响应失败", err.Error(), err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", apperror.New("E302", fmt.Sprintf("API 返回错误状态码: %d", resp.StatusCode), string(body), nil)
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", apperror.New("E302", "解析响应失败", err.Error(), err)
	}

	if len(genResp.Data) == 0 {
		return "", apperror.New("E303", "API 未返回图片", "", nil)
	}

	// 返回图片URL
	return genResp.Data[0].URL, nil
}

// DownloadImage 下载图片到本地
func (c *ImageGenerationClient) DownloadImage(ctx context.Context, imageURL, outputPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return apperror.New("E304", "创建下载请求失败", err.Error(), err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperror.New("E304", "下载图片失败", err.Error(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apperror.New("E304", fmt.Sprintf("下载图片失败，状态码: %d", resp.StatusCode), "", nil)
	}

	// 创建输出目录
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return apperror.New("E304", "创建输出目录失败", err.Error(), err)
	}

	// 保存图片
	outFile, err := os.Create(outputPath)
	if err != nil {
		return apperror.New("E304", "创建输出文件失败", err.Error(), err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return apperror.New("E304", "保存图片失败", err.Error(), err)
	}

	return nil
}
