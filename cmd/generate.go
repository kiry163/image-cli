package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/kiry163/image-cli/internal/ai"
	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `generate "prompt"`,
		Short: "AI图像生成",
		Long:  "使用智谱AI CogView模型根据文本描述生成图像",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]

			model, _ := cmd.Flags().GetString("model")
			size, _ := cmd.Flags().GetString("size")
			quality, _ := cmd.Flags().GetString("quality")
			output, _ := cmd.Flags().GetString("output")

			cfg := CurrentConfig()

			// 检查API Key是否配置
			if cfg.ImageGeneration.APIKey == "" {
				return apperror.New("E305", "图像生成 API Key 未配置", "请在配置文件中设置 image_generation.api_key 或使用环境变量 IMAGE_GENERATION_API_KEY", nil)
			}

			// 使用默认值
			if model == "" {
				model = cfg.ImageGeneration.DefaultModel
			}
			if size == "" {
				size = cfg.ImageGeneration.DefaultSize
			}
			if quality == "" {
				quality = cfg.ImageGeneration.DefaultQuality
			}
			if output == "" {
				output = filepath.Join(cfg.Base.OutputDir, fmt.Sprintf("generated_%d.png", time.Now().Unix()))
			}

			// 创建客户端
			client, err := ai.NewImageGenerationClient(cfg.ImageGeneration.APIKey, cfg.ImageGeneration.BaseURL)
			if err != nil {
				return err
			}

			// 执行生成
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			if !quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "正在生成图像...\n")
				fmt.Fprintf(cmd.OutOrStdout(), "提示词: %s\n", prompt)
				fmt.Fprintf(cmd.OutOrStdout(), "模型: %s, 尺寸: %s, 质量: %s\n", model, size, quality)
			}

			imageURL, err := client.Generate(ctx, prompt, ai.GenerateOptions{
				Model:   model,
				Size:    size,
				Quality: quality,
			})
			if err != nil {
				return err
			}

			if !quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "图片已生成，正在下载...\n")
			}

			// 下载图片
			if err := client.DownloadImage(ctx, imageURL, output); err != nil {
				return err
			}

			if !quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "图片已保存: %s\n", output)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), output)
			}

			return nil
		},
	}

	cmd.Flags().StringP("model", "m", "", "模型: cogview-3-flash(免费), glm-image, cogview-4")
	cmd.Flags().StringP("size", "s", "", "图片尺寸: 1024x1024, 768x1344, 864x1152 等")
	cmd.Flags().StringP("quality", "q", "", "图片质量: standard(默认), hd")
	cmd.Flags().StringP("output", "o", "", "输出文件路径")

	return cmd
}
