package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/kiry163/image-cli/pkg/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
	}
	cmd.AddCommand(newConfigInitCmd())
	return cmd
}

func newConfigInitCmd() *cobra.Command {
	var overwrite bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "生成默认配置文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := config.ConfigPath("")
			if path == "" {
				return apperror.ConfigError("无法确定配置文件路径", nil)
			}
			if _, err := os.Stat(path); err == nil && !overwrite {
				return apperror.OutputExists("配置文件已存在: " + path)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return apperror.ConfigError("无法创建配置目录", err)
			}
			if err := os.WriteFile(path, []byte(defaultConfigYAML), 0o644); err != nil {
				return apperror.ConfigError("无法写入配置文件", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已生成: %s\n", path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "覆盖已存在配置文件")
	return cmd
}

const defaultConfigYAML = "" +
	"# ImageCLI 配置\n" +
	"\n" +
	"# 基础设置\n" +
	"base:\n" +
	"  output_dir: ./output\n" +
	"  overwrite: false\n" +
	"  keep_temp: false\n" +
	"  recursive: true\n" +
	"  conflict: skip\n" +
	"\n" +
	"# 压缩设置\n" +
	"compress:\n" +
	"  default_quality: 85\n" +
	"  max_width: 4096\n" +
	"  max_height: 4096\n" +
	"\n" +
	"# 水印设置\n" +
	"watermark:\n" +
	"  default_opacity: 0.5\n" +
	"  default_scale: 0.2\n" +
	"  default_gravity: southeast\n" +
	"  default_offset_x: 0\n" +
	"  default_offset_y: 0\n" +
	"  default_font_size: 24\n" +
	"  default_font: \"\"\n" +
	"  default_font_file: \"\"\n" +
	"  default_color: white\n" +
	"  default_stroke_color: \"\"\n" +
	"  default_stroke_width: 0\n" +
	"  default_background: none\n" +
	"  default_stroke_mode: circle\n" +
	"\n" +
	"# AI 模型配置\n" +
	"ai:\n" +
	"  default_model: gpt-4o\n" +
	"\n" +
	"  output:\n" +
	"    default_format: \"\"\n" +
	"    remove_bg_format: png\n" +
	"\n" +
	"  models:\n" +
	"    gpt-4o:\n" +
	"      provider: openai\n" +
	"      api_key_env: OPENAI_API_KEY\n" +
	"      endpoint: https://api.openai.com/v1\n" +
	"\n" +
	"    claude-3-5-sonnet:\n" +
	"      provider: anthropic\n" +
	"      api_key_env: ANTHROPIC_API_KEY\n" +
	"      endpoint: https://api.anthropic.com\n" +
	"\n" +
	"    gemini-1.5-pro:\n" +
	"      provider: google\n" +
	"      api_key_env: GOOGLE_API_KEY\n" +
	"      endpoint: https://generativelanguage.googleapis.com/v1\n" +
	"\n" +
	"# OCR 文字识别配置\n" +
	"ocr:\n" +
	"  api_key: \"\"\n" +
	"  base_url: \"https://www.dmxapi.cn/v1\"\n" +
	"  model: \"DeepSeek-OCR\"\n" +
	"  default_mode: \"free\"\n" +
	"\n" +
	"# AI图像生成配置\n" +
	"image_generation:\n" +
	"  api_key: \"\"\n" +
	"  base_url: \"https://open.bigmodel.cn/api/paas/v4\"\n" +
	"  default_model: \"cogview-3-flash\"\n" +
	"  default_size: \"1024x1024\"\n" +
	"  default_quality: \"standard\"\n" +
	"\n" +
	"# AI图片识别配置\n" +
	"vision:\n" +
	"  api_key: \"\"\n" +
	"  base_url: \"https://open.bigmodel.cn/api/paas/v4\"\n" +
	"  default_model: \"glm-4v-flash\"\n" +
	"  default_prompt: \"请描述这张图片的内容\"\n" +
	"\n" +
	"# 日志设置\n" +
	"logging:\n" +
	"  level: info\n" +
	"  format: json\n"
