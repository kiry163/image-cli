package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/spf13/viper"
)

type Config struct {
	Base            BaseConfig            `mapstructure:"base"`
	Compress        CompressConfig        `mapstructure:"compress"`
	Watermark       WatermarkConfig       `mapstructure:"watermark"`
	OCR             OCRConfig             `mapstructure:"ocr"`
	ImageGeneration ImageGenerationConfig `mapstructure:"image_generation"`
	Vision          VisionConfig          `mapstructure:"vision"`
	AI              AIConfig              `mapstructure:"ai"`
	Logging         LoggingConfig         `mapstructure:"logging"`
}

type BaseConfig struct {
	OutputDir string `mapstructure:"output_dir"`
	Overwrite bool   `mapstructure:"overwrite"`
	KeepTemp  bool   `mapstructure:"keep_temp"`
	Recursive bool   `mapstructure:"recursive"`
	Conflict  string `mapstructure:"conflict"`
}

type CompressConfig struct {
	DefaultQuality int `mapstructure:"default_quality"`
	MaxWidth       int `mapstructure:"max_width"`
	MaxHeight      int `mapstructure:"max_height"`
}

type WatermarkConfig struct {
	DefaultOpacity     float64 `mapstructure:"default_opacity"`
	DefaultScale       float64 `mapstructure:"default_scale"`
	DefaultGravity     string  `mapstructure:"default_gravity"`
	DefaultOffsetX     int     `mapstructure:"default_offset_x"`
	DefaultOffsetY     int     `mapstructure:"default_offset_y"`
	DefaultFontSize    int     `mapstructure:"default_font_size"`
	DefaultFont        string  `mapstructure:"default_font"`
	DefaultFontFile    string  `mapstructure:"default_font_file"`
	DefaultColor       string  `mapstructure:"default_color"`
	DefaultStrokeColor string  `mapstructure:"default_stroke_color"`
	DefaultStrokeWidth int     `mapstructure:"default_stroke_width"`
	DefaultBackground  string  `mapstructure:"default_background"`
	DefaultStrokeMode  string  `mapstructure:"default_stroke_mode"`
}

type OCRConfig struct {
	APIKey      string `mapstructure:"api_key"`
	BaseURL     string `mapstructure:"base_url"`
	Model       string `mapstructure:"model"`
	DefaultMode string `mapstructure:"default_mode"`
}

type ImageGenerationConfig struct {
	APIKey         string `mapstructure:"api_key"`
	BaseURL        string `mapstructure:"base_url"`
	DefaultModel   string `mapstructure:"default_model"`
	DefaultSize    string `mapstructure:"default_size"`
	DefaultQuality string `mapstructure:"default_quality"`
}

type VisionConfig struct {
	APIKey        string `mapstructure:"api_key"`
	BaseURL       string `mapstructure:"base_url"`
	DefaultModel  string `mapstructure:"default_model"`
	DefaultPrompt string `mapstructure:"default_prompt"`
}

type AIConfig struct {
	DefaultModel string             `mapstructure:"default_model"`
	Output       AIOutputConfig     `mapstructure:"output"`
	Models       map[string]AIModel `mapstructure:"models"`
}

type AIOutputConfig struct {
	DefaultFormat  string `mapstructure:"default_format"`
	RemoveBGFormat string `mapstructure:"remove_bg_format"`
}

type AIModel struct {
	Provider  string `mapstructure:"provider"`
	APIKeyEnv string `mapstructure:"api_key_env"`
	Endpoint  string `mapstructure:"endpoint"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func NewViper() *viper.Viper {
	v := viper.New()
	v.SetDefault("base.output_dir", "./output")
	v.SetDefault("base.overwrite", false)
	v.SetDefault("base.keep_temp", false)
	v.SetDefault("base.recursive", true)
	v.SetDefault("base.conflict", "skip")

	v.SetDefault("compress.default_quality", 85)
	v.SetDefault("compress.max_width", 4096)
	v.SetDefault("compress.max_height", 4096)

	v.SetDefault("watermark.default_opacity", 0.5)
	v.SetDefault("watermark.default_scale", 0.2)
	v.SetDefault("watermark.default_gravity", "southeast")
	v.SetDefault("watermark.default_offset_x", 0)
	v.SetDefault("watermark.default_offset_y", 0)
	v.SetDefault("watermark.default_font_size", 24)
	v.SetDefault("watermark.default_font", "")
	v.SetDefault("watermark.default_font_file", "")
	v.SetDefault("watermark.default_color", "white")
	v.SetDefault("watermark.default_stroke_color", "")
	v.SetDefault("watermark.default_stroke_width", 0)
	v.SetDefault("watermark.default_background", "none")
	v.SetDefault("watermark.default_stroke_mode", "circle")

	v.SetDefault("ocr.api_key", "")
	v.SetDefault("ocr.base_url", "https://www.dmxapi.cn/v1")
	v.SetDefault("ocr.model", "DeepSeek-OCR")
	v.SetDefault("ocr.default_mode", "free")

	v.SetDefault("image_generation.api_key", "")
	v.SetDefault("image_generation.base_url", "https://open.bigmodel.cn/api/paas/v4")
	v.SetDefault("image_generation.default_model", "cogview-3-flash")
	v.SetDefault("image_generation.default_size", "1024x1024")
	v.SetDefault("image_generation.default_quality", "standard")

	v.SetDefault("vision.api_key", "")
	v.SetDefault("vision.base_url", "https://open.bigmodel.cn/api/paas/v4")
	v.SetDefault("vision.default_model", "glm-4v-flash")
	v.SetDefault("vision.default_prompt", "请描述这张图片的内容")

	v.SetDefault("ai.default_model", "gpt-4o")
	v.SetDefault("ai.output.default_format", "")
	v.SetDefault("ai.output.remove_bg_format", "png")

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	v.SetEnvPrefix("IMAGE_CLI")
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	v.BindEnv("base.output_dir", "IMAGE_CLI_OUTPUT")
	v.BindEnv("base.recursive", "IMAGE_CLI_RECURSIVE")
	v.BindEnv("ocr.api_key", "OCR_API_KEY")
	v.BindEnv("ocr.base_url", "OCR_BASE_URL")
	v.BindEnv("image_generation.api_key", "IMAGE_GENERATION_API_KEY")
	v.BindEnv("image_generation.base_url", "IMAGE_GENERATION_BASE_URL")
	v.BindEnv("vision.api_key", "IMAGE_VISION_API_KEY")
	v.BindEnv("vision.base_url", "IMAGE_VISION_BASE_URL")
	v.AutomaticEnv()

	return v
}

func Load(v *viper.Viper, path string, explicit bool) error {
	if path == "" {
		return apperror.ConfigError("配置路径为空", nil)
	}
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) && !explicit {
			return nil
		}
		return apperror.ConfigError("无法读取配置文件", err)
	}
	return nil
}

func FromViper(v *viper.Viper) (Config, error) {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, apperror.ConfigError("配置解析失败", err)
	}
	if cfg.Base.Conflict == "" {
		cfg.Base.Conflict = "skip"
	}
	switch cfg.Base.Conflict {
	case "skip", "overwrite", "rename":
	default:
		return Config{}, apperror.ConfigError("冲突策略无效", nil)
	}
	return cfg, nil
}

func ConfigPath(flagValue string) (string, bool) {
	if flagValue != "" {
		return flagValue, true
	}
	if env := os.Getenv("IMAGE_CLI_CONFIG"); env != "" {
		return env, true
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "./config.yaml", false
	}
	return filepath.Join(home, ".config", "image-cli", "config.yaml"), false
}
