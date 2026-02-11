package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kiry163/image-cli/internal/batch"
	"github.com/kiry163/image-cli/internal/core"
	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		newConfigCmd(),
		newVersionCmd(),
		newFormatsCmd(),
		newInfoCmd(),
		newConvertCmd(),
		newCompressCmd(),
		newResizeCmd(),
		newRotateCmd(),
		newWatermarkCmd(),
		newBatchCmd(),
		newRemoveWatermarkCmd(),
		newRemoveBgCmd(),
		newEnhanceCmd(),
		newStyleTransferCmd(),
		newOcrCmd(),
		newGenerateCmd(),
		newRecognizeCmd(),
	)
}

func newNotImplementedCmd(use, short string, ai bool) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ai {
				return apperror.AINotImplemented()
			}
			return apperror.NotImplemented(short)
		},
	}
}

func newConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert <input> <output>",
		Short: "格式转换",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			quality, _ := cmd.Flags().GetInt("quality")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			icoSizesRaw, _ := cmd.Flags().GetString("ico-sizes")
			icoSizes, err := core.ParseICOSizes(icoSizesRaw)
			if err != nil {
				return err
			}
			cfg := CurrentConfig()
			outPath, err := core.Convert(args[0], args[1], core.ConvertOptions{
				Format:    format,
				Quality:   quality,
				Overwrite: overwrite,
				Conflict:  cfg.Base.Conflict,
				ICOSizes:  icoSizes,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "输出: %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().StringP("format", "f", "", "输出格式")
	cmd.Flags().IntP("quality", "q", 85, "质量 (1-100)")
	cmd.Flags().Bool("overwrite", false, "覆盖已存在文件")
	cmd.Flags().String("ico-sizes", "", "ICO 尺寸列表 (如 256,128,64)")
	return cmd
}

func newCompressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compress <input>",
		Short: "图像压缩",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			quality, _ := cmd.Flags().GetInt("quality")
			maxSize, _ := cmd.Flags().GetString("max-size")
			output, _ := cmd.Flags().GetString("output")
			aggressive, _ := cmd.Flags().GetBool("aggressive")
			cfg := CurrentConfig()
			if output == "" {
				output = cfg.Base.OutputDir
			}
			maxSizeBytes, err := core.ParseSizeBytes(maxSize)
			if err != nil {
				return err
			}
			outPath, err := core.Compress(args[0], output, core.CompressOptions{
				Quality:        quality,
				MaxSizeBytes:   maxSizeBytes,
				Aggressive:     aggressive,
				Conflict:       cfg.Base.Conflict,
				DefaultQuality: cfg.Compress.DefaultQuality,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "输出: %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().IntP("quality", "Q", 0, "JPEG/WebP 质量 (1-100)")
	cmd.Flags().String("max-size", "", "最大文件大小")
	cmd.Flags().StringP("output", "o", "", "输出路径")
	cmd.Flags().Bool("aggressive", false, "激进压缩")
	return cmd
}

func newResizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resize <input> <output>",
		Short: "尺寸调整",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			width, _ := cmd.Flags().GetString("width")
			height, _ := cmd.Flags().GetString("height")
			fit, _ := cmd.Flags().GetString("fit")
			withoutEnlargement, _ := cmd.Flags().GetBool("without-enlargement")
			keepRatio, _ := cmd.Flags().GetBool("keep-ratio")
			cfg := CurrentConfig()
			outPath, err := core.Resize(args[0], args[1], core.ResizeOptions{
				Width:              width,
				Height:             height,
				Fit:                fit,
				WithoutEnlargement: withoutEnlargement,
				KeepRatio:          keepRatio,
				Conflict:           cfg.Base.Conflict,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "输出: %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().StringP("width", "w", "", "宽度")
	cmd.Flags().String("height", "", "高度")
	cmd.Flags().StringP("fit", "f", "", "适应模式")
	cmd.Flags().Bool("without-enlargement", true, "不放大")
	cmd.Flags().Bool("keep-ratio", true, "保持比例")
	return cmd
}

func newRotateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate <input> <output>",
		Short: "旋转翻转",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			degrees, _ := cmd.Flags().GetInt("degrees")
			flip, _ := cmd.Flags().GetBool("flip")
			flop, _ := cmd.Flags().GetBool("flop")
			cfg := CurrentConfig()
			outPath, err := core.Rotate(args[0], args[1], core.RotateOptions{
				Degrees:  degrees,
				Flip:     flip,
				Flop:     flop,
				Conflict: cfg.Base.Conflict,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "输出: %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().IntP("degrees", "d", 0, "旋转角度")
	cmd.Flags().Bool("flip", false, "水平翻转")
	cmd.Flags().Bool("flop", false, "垂直翻转")
	return cmd
}

func newWatermarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watermark <input> <logo> <output> | watermark <input> <output> --text \"...\"",
		Short: "添加水印",
		Args: func(cmd *cobra.Command, args []string) error {
			text, _ := cmd.Flags().GetString("text")
			if text != "" {
				if len(args) != 2 {
					return apperror.InvalidArgument("文本水印需要输入与输出参数", nil)
				}
				return nil
			}
			if len(args) != 3 {
				return apperror.InvalidArgument("图片水印需要输入、logo 与输出参数", nil)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := CurrentConfig()
			gravity, _ := cmd.Flags().GetString("gravity")
			opacity, _ := cmd.Flags().GetFloat64("opacity")
			scale, _ := cmd.Flags().GetFloat64("scale")
			offsetX, _ := cmd.Flags().GetInt("offset-x")
			offsetY, _ := cmd.Flags().GetInt("offset-y")
			text, _ := cmd.Flags().GetString("text")
			fontSize, _ := cmd.Flags().GetInt("font-size")
			font, _ := cmd.Flags().GetString("font")
			fontFile, _ := cmd.Flags().GetString("font-file")
			color, _ := cmd.Flags().GetString("color")
			strokeColor, _ := cmd.Flags().GetString("stroke-color")
			strokeWidth, _ := cmd.Flags().GetInt("stroke-width")
			background, _ := cmd.Flags().GetString("background")
			strokeMode, _ := cmd.Flags().GetString("stroke-mode")
			if opacity <= 0 {
				opacity = cfg.Watermark.DefaultOpacity
			}
			if scale <= 0 {
				scale = cfg.Watermark.DefaultScale
			}
			if gravity == "" {
				gravity = cfg.Watermark.DefaultGravity
			}
			if offsetX == 0 {
				offsetX = cfg.Watermark.DefaultOffsetX
			}
			if offsetY == 0 {
				offsetY = cfg.Watermark.DefaultOffsetY
			}
			if fontSize <= 0 {
				fontSize = cfg.Watermark.DefaultFontSize
			}
			if font == "" {
				font = cfg.Watermark.DefaultFont
			}
			if fontFile == "" {
				fontFile = cfg.Watermark.DefaultFontFile
			}
			if color == "" {
				color = cfg.Watermark.DefaultColor
			}
			if strokeColor == "" {
				strokeColor = cfg.Watermark.DefaultStrokeColor
			}
			if strokeWidth == 0 {
				strokeWidth = cfg.Watermark.DefaultStrokeWidth
			}
			if background == "" {
				background = cfg.Watermark.DefaultBackground
			}
			if strokeMode == "" {
				strokeMode = cfg.Watermark.DefaultStrokeMode
			}
			input := args[0]
			var output string
			logo := ""
			if text != "" {
				output = args[1]
			} else {
				logo = args[1]
				output = args[2]
			}
			outPath, err := core.Watermark(input, output, core.WatermarkOptions{
				LogoPath:    logo,
				Text:        text,
				Opacity:     opacity,
				Scale:       scale,
				Gravity:     gravity,
				OffsetX:     offsetX,
				OffsetY:     offsetY,
				FontSize:    fontSize,
				Font:        font,
				FontFile:    fontFile,
				Color:       color,
				StrokeColor: strokeColor,
				StrokeWidth: strokeWidth,
				Background:  background,
				StrokeMode:  strokeMode,
				Conflict:    cfg.Base.Conflict,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "输出: %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().StringP("gravity", "g", "", "位置")
	cmd.Flags().Float64P("opacity", "o", 0, "透明度")
	cmd.Flags().Float64P("scale", "s", 0, "缩放比例")
	cmd.Flags().Int("offset-x", 0, "水平偏移(px)")
	cmd.Flags().Int("offset-y", 0, "垂直偏移(px)")
	cmd.Flags().String("text", "", "文字水印")
	cmd.Flags().Int("font-size", 0, "文字水印字号(px)")
	cmd.Flags().String("font", "", "文字水印字体")
	cmd.Flags().String("font-file", "", "文字水印字体文件")
	cmd.Flags().String("color", "", "文字水印颜色")
	cmd.Flags().String("stroke-color", "", "文字水印描边颜色")
	cmd.Flags().Int("stroke-width", 0, "文字水印描边宽度(px)")
	cmd.Flags().String("background", "", "文字水印背景色")
	cmd.Flags().String("stroke-mode", "", "描边模式: circle|8dir")
	return cmd
}

func newBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch <command> <pattern>",
		Short: "批量处理",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := CurrentConfig()
			sub := args[0]
			pattern := args[1]
			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = cfg.Base.OutputDir
			}
			collected, err := batch.Collect(pattern, cfg.Base.Recursive)
			if err != nil {
				return err
			}
			total := len(collected.Files)
			if !quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "开始: %d\n", total)
			}
			success := 0
			failed := 0
			errOut := cmd.ErrOrStderr()
			for _, input := range collected.Files {
				if verbose && !quiet {
					fmt.Fprintf(cmd.OutOrStdout(), "处理: %s\n", input)
				}
				outDir := output
				if rel, err := filepath.Rel(collected.BaseDir, input); err == nil && !strings.HasPrefix(rel, "..") {
					outDir = filepath.Join(output, filepath.Dir(rel))
				}
				switch sub {
				case "convert":
					format, _ := cmd.Flags().GetString("to")
					quality, _ := cmd.Flags().GetInt("quality")
					_, err = core.Convert(input, outDir, core.ConvertOptions{
						Format:    format,
						Quality:   quality,
						Overwrite: false,
						Conflict:  cfg.Base.Conflict,
					})
				case "compress":
					quality, _ := cmd.Flags().GetInt("quality")
					maxSize, _ := cmd.Flags().GetString("max-size")
					aggressive, _ := cmd.Flags().GetBool("aggressive")
					maxSizeBytes, parseErr := core.ParseSizeBytes(maxSize)
					if parseErr != nil {
						return parseErr
					}
					_, err = core.Compress(input, outDir, core.CompressOptions{
						Quality:        quality,
						MaxSizeBytes:   maxSizeBytes,
						Aggressive:     aggressive,
						Conflict:       cfg.Base.Conflict,
						DefaultQuality: cfg.Compress.DefaultQuality,
					})
				case "resize":
					width, _ := cmd.Flags().GetString("width")
					height, _ := cmd.Flags().GetString("height")
					fit, _ := cmd.Flags().GetString("fit")
					withoutEnlargement, _ := cmd.Flags().GetBool("without-enlargement")
					keepRatio, _ := cmd.Flags().GetBool("keep-ratio")
					_, err = core.Resize(input, outDir, core.ResizeOptions{
						Width:              width,
						Height:             height,
						Fit:                fit,
						WithoutEnlargement: withoutEnlargement,
						KeepRatio:          keepRatio,
						Conflict:           cfg.Base.Conflict,
					})
				case "rotate":
					degrees, _ := cmd.Flags().GetInt("degrees")
					flip, _ := cmd.Flags().GetBool("flip")
					flop, _ := cmd.Flags().GetBool("flop")
					_, err = core.Rotate(input, outDir, core.RotateOptions{
						Degrees:  degrees,
						Flip:     flip,
						Flop:     flop,
						Conflict: cfg.Base.Conflict,
					})
				case "watermark":
					gravity, _ := cmd.Flags().GetString("gravity")
					opacity, _ := cmd.Flags().GetFloat64("opacity")
					scale, _ := cmd.Flags().GetFloat64("scale")
					offsetX, _ := cmd.Flags().GetInt("offset-x")
					offsetY, _ := cmd.Flags().GetInt("offset-y")
					text, _ := cmd.Flags().GetString("text")
					fontSize, _ := cmd.Flags().GetInt("font-size")
					font, _ := cmd.Flags().GetString("font")
					fontFile, _ := cmd.Flags().GetString("font-file")
					color, _ := cmd.Flags().GetString("color")
					strokeColor, _ := cmd.Flags().GetString("stroke-color")
					strokeWidth, _ := cmd.Flags().GetInt("stroke-width")
					background, _ := cmd.Flags().GetString("background")
					strokeMode, _ := cmd.Flags().GetString("stroke-mode")
					logo, _ := cmd.Flags().GetString("logo")
					if text == "" && logo == "" {
						return apperror.InvalidArgument("批量水印需要 --logo 或 --text", nil)
					}
					if text != "" && logo != "" {
						return apperror.InvalidArgument("--logo 与 --text 不可同时使用", nil)
					}
					if opacity <= 0 {
						opacity = cfg.Watermark.DefaultOpacity
					}
					if scale <= 0 {
						scale = cfg.Watermark.DefaultScale
					}
					if gravity == "" {
						gravity = cfg.Watermark.DefaultGravity
					}
					if offsetX == 0 {
						offsetX = cfg.Watermark.DefaultOffsetX
					}
					if offsetY == 0 {
						offsetY = cfg.Watermark.DefaultOffsetY
					}
					if fontSize <= 0 {
						fontSize = cfg.Watermark.DefaultFontSize
					}
					if font == "" {
						font = cfg.Watermark.DefaultFont
					}
					if fontFile == "" {
						fontFile = cfg.Watermark.DefaultFontFile
					}
					if color == "" {
						color = cfg.Watermark.DefaultColor
					}
					if strokeColor == "" {
						strokeColor = cfg.Watermark.DefaultStrokeColor
					}
					if strokeWidth == 0 {
						strokeWidth = cfg.Watermark.DefaultStrokeWidth
					}
					if background == "" {
						background = cfg.Watermark.DefaultBackground
					}
					if strokeMode == "" {
						strokeMode = cfg.Watermark.DefaultStrokeMode
					}
					_, err = core.Watermark(input, outDir, core.WatermarkOptions{
						LogoPath:    logo,
						Text:        text,
						Opacity:     opacity,
						Scale:       scale,
						Gravity:     gravity,
						OffsetX:     offsetX,
						OffsetY:     offsetY,
						FontSize:    fontSize,
						Font:        font,
						FontFile:    fontFile,
						Color:       color,
						StrokeColor: strokeColor,
						StrokeWidth: strokeWidth,
						Background:  background,
						StrokeMode:  strokeMode,
						Conflict:    cfg.Base.Conflict,
					})
				default:
					return apperror.InvalidArgument("不支持的批量命令", nil)
				}
				if err != nil {
					failed++
					if !quiet {
						fmt.Fprintf(errOut, "失败: %s\n", input)
						WriteError(errOut, err)
					}
					continue
				}
				success++
			}
			if !quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "完成: %d 成功, %d 失败\n", success, failed)
			}
			if failed > 0 {
				return apperror.BatchFailed(fmt.Sprintf("失败 %d 个文件", failed))
			}
			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "", "输出路径")
	cmd.Flags().String("to", "", "目标格式")
	cmd.Flags().IntP("quality", "q", 85, "质量 (1-100)")
	cmd.Flags().String("max-size", "", "最大文件大小")
	cmd.Flags().Bool("aggressive", false, "激进压缩")
	cmd.Flags().String("logo", "", "水印图像")
	cmd.Flags().String("text", "", "文字水印")
	cmd.Flags().Int("font-size", 0, "文字水印字号(px)")
	cmd.Flags().String("font", "", "文字水印字体")
	cmd.Flags().String("font-file", "", "文字水印字体文件")
	cmd.Flags().String("color", "", "文字水印颜色")
	cmd.Flags().String("stroke-color", "", "文字水印描边颜色")
	cmd.Flags().Int("stroke-width", 0, "文字水印描边宽度(px)")
	cmd.Flags().String("background", "", "文字水印背景色")
	cmd.Flags().String("stroke-mode", "", "描边模式: circle|8dir")
	cmd.Flags().String("width", "", "宽度")
	cmd.Flags().String("height", "", "高度")
	cmd.Flags().String("fit", "", "适应模式")
	cmd.Flags().Bool("without-enlargement", true, "不放大")
	cmd.Flags().Bool("keep-ratio", true, "保持比例")
	cmd.Flags().IntP("degrees", "d", 0, "旋转角度")
	cmd.Flags().Bool("flip", false, "水平翻转")
	cmd.Flags().Bool("flop", false, "垂直翻转")
	cmd.Flags().StringP("gravity", "g", "", "位置")
	cmd.Flags().Float64("opacity", 0, "透明度")
	cmd.Flags().Float64P("scale", "s", 0, "缩放比例")
	cmd.Flags().Int("offset-x", 0, "水平偏移(px)")
	cmd.Flags().Int("offset-y", 0, "垂直偏移(px)")
	return cmd
}

func newRemoveWatermarkCmd() *cobra.Command {
	cmd := newNotImplementedCmd("remove-watermark <input>", "去除水印", true)
	cmd.Flags().StringP("output", "o", "", "输出路径")
	cmd.Flags().StringP("model", "m", "", "使用模型")
	cmd.Flags().String("api-key", "", "API Key")
	cmd.Flags().String("format", "", "输出格式")
	return cmd
}

func newRemoveBgCmd() *cobra.Command {
	cmd := newNotImplementedCmd("remove-bg <input>", "智能抠图", true)
	cmd.Flags().StringP("output", "o", "", "输出路径")
	cmd.Flags().StringP("model", "m", "", "使用模型")
	cmd.Flags().Bool("matte", false, "保留边缘细节")
	cmd.Flags().String("format", "", "输出格式")
	return cmd
}

func newEnhanceCmd() *cobra.Command {
	cmd := newNotImplementedCmd("enhance <input>", "AI 图像增强", true)
	cmd.Flags().IntP("scale", "s", 2, "放大倍数")
	cmd.Flags().StringP("model", "m", "", "超分辨率模型")
	cmd.Flags().Bool("denoise", false, "降噪")
	cmd.Flags().Bool("sharpen", false, "锐化")
	cmd.Flags().String("format", "", "输出格式")
	return cmd
}

func newStyleTransferCmd() *cobra.Command {
	cmd := newNotImplementedCmd("style-transfer <input>", "风格迁移", true)
	cmd.Flags().String("style", "", "风格名称")
	cmd.Flags().StringP("output", "o", "", "输出路径")
	cmd.Flags().String("format", "", "输出格式")
	return cmd
}
