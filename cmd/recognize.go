package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kiry163/image-cli/internal/ai"
	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/spf13/cobra"
)

func newRecognizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recognize <image>",
		Short: "AI图片识别",
		Long:  "使用智谱AI GLM-4V模型对图片进行视觉理解和分析",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			imagePath := args[0]

			model, _ := cmd.Flags().GetString("model")
			prompt, _ := cmd.Flags().GetString("prompt")
			output, _ := cmd.Flags().GetString("output")

			cfg := CurrentConfig()

			// 检查API Key是否配置
			if cfg.Vision.APIKey == "" {
				return apperror.New("E405", "视觉理解 API Key 未配置", "请在配置文件中设置 vision.api_key 或使用环境变量 IMAGE_VISION_API_KEY", nil)
			}

			// 使用默认值
			if model == "" {
				model = cfg.Vision.DefaultModel
			}
			if prompt == "" {
				prompt = cfg.Vision.DefaultPrompt
			}

			// 创建客户端
			client, err := ai.NewVisionClient(cfg.Vision.APIKey, cfg.Vision.BaseURL)
			if err != nil {
				return err
			}

			// 执行分析
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			if !quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "正在分析图片...\n")
				fmt.Fprintf(cmd.OutOrStdout(), "模型: %s\n", model)
				fmt.Fprintf(cmd.OutOrStdout(), "提示: %s\n", prompt)
			}

			result, err := client.Analyze(ctx, imagePath, ai.VisionOptions{
				Model:  model,
				Prompt: prompt,
			})
			if err != nil {
				return err
			}

			// 输出结果
			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0644); err != nil {
					return apperror.New("E103", "无法写入输出文件", err.Error(), err)
				}
				if !quiet {
					fmt.Fprintf(cmd.OutOrStdout(), "结果已保存: %s\n", output)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), result)
			}

			return nil
		},
	}

	cmd.Flags().StringP("model", "m", "", "模型: glm-4v-flash(免费), glm-4.6v")
	cmd.Flags().StringP("prompt", "p", "", "分析提示词/问题")
	cmd.Flags().StringP("output", "o", "", "输出文件路径")

	return cmd
}
