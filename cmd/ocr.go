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

func newOcrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ocr <input>",
		Short: "OCR文字识别",
		Long:  "使用DeepSeek OCR API对图片进行文字识别",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]
			mode, _ := cmd.Flags().GetString("mode")
			output, _ := cmd.Flags().GetString("output")

			cfg := CurrentConfig()

			// 检查API Key是否配置
			if cfg.OCR.APIKey == "" {
				return apperror.New("E202", "OCR API Key 未配置", "请在配置文件中设置 ocr.api_key 或使用环境变量 OCR_API_KEY", nil)
			}

			// 如果mode未指定，使用配置默认值
			if mode == "" {
				mode = cfg.OCR.DefaultMode
			}

			// 创建OCR客户端
			client, err := ai.NewOCRClient(cfg.OCR.APIKey, cfg.OCR.BaseURL, cfg.OCR.Model)
			if err != nil {
				return err
			}

			// 执行OCR识别
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			if !quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "正在进行OCR识别: %s (模式: %s)\n", input, mode)
			}

			result, err := client.Recognize(ctx, input, ai.OCROptions{Mode: mode})
			if err != nil {
				return err
			}

			// 输出结果
			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0644); err != nil {
					return apperror.New("E103", "无法写入输出文件", err.Error(), err)
				}
				if !quiet {
					fmt.Fprintf(cmd.OutOrStdout(), "结果已保存至: %s\n", output)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), result)
			}

			return nil
		},
	}

	cmd.Flags().StringP("mode", "m", "", "识别模式: free(默认), markdown, text, figure, detail")
	cmd.Flags().StringP("output", "o", "", "输出文件路径（默认输出到控制台）")

	return cmd
}
