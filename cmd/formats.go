package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/internal/core"
	"github.com/spf13/cobra"
)

type formatSupport struct {
	Name string
	Type bimg.ImageType
}

func newFormatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "formats",
		Short: "查看支持的格式与转换",
		RunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			from = core.NormalizeFormat(from)
			to = core.NormalizeFormat(to)
			inputFormats := supportedFormats(true)
			outputFormats := supportedFormats(false)
			if from != "" {
				inputFormats = filterFormats(inputFormats, from)
			}
			if to != "" {
				outputFormats = filterFormats(outputFormats, to)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "输入格式: %s\n", strings.Join(inputFormats, ", "))
			fmt.Fprintf(cmd.OutOrStdout(), "输出格式: %s\n", strings.Join(outputFormats, ", "))
			fmt.Fprintln(cmd.OutOrStdout(), "转换支持:")
			pairs := buildPairs(inputFormats, outputFormats)
			for _, pair := range pairs {
				fmt.Fprintln(cmd.OutOrStdout(), pair)
			}
			return nil
		},
	}
	cmd.Flags().String("from", "", "输入格式过滤")
	cmd.Flags().String("to", "", "输出格式过滤")
	return cmd
}

func supportedFormats(input bool) []string {
	formats := []formatSupport{
		{Name: "jpg", Type: bimg.JPEG},
		{Name: "png", Type: bimg.PNG},
		{Name: "webp", Type: bimg.WEBP},
		{Name: "gif", Type: bimg.GIF},
		{Name: "tiff", Type: bimg.TIFF},
		{Name: "pdf", Type: bimg.PDF},
		{Name: "heif", Type: bimg.HEIF},
		{Name: "avif", Type: bimg.AVIF},
		{Name: "svg", Type: bimg.SVG},
	}
	result := make([]string, 0, len(formats))
	for _, item := range formats {
		name := core.NormalizeFormat(item.Name)
		if input {
			if bimg.IsTypeSupported(item.Type) {
				result = append(result, name)
			}
			continue
		}
		if bimg.IsTypeSupportedSave(item.Type) {
			result = append(result, name)
		}
	}
	if !input && core.HasImageMagick() && bimg.IsTypeSupportedSave(bimg.PNG) {
		result = append(result, "ico")
	}
	sort.Strings(result)
	return result
}

func buildPairs(inputs []string, outputs []string) []string {
	if len(inputs) == 0 || len(outputs) == 0 {
		return []string{}
	}
	pairs := make([]string, 0, len(inputs)*len(outputs))
	for _, in := range inputs {
		for _, out := range outputs {
			pairs = append(pairs, fmt.Sprintf("%s -> %s", in, out))
		}
	}
	return pairs
}

func filterFormats(formats []string, value string) []string {
	if value == "" {
		return formats
	}
	filtered := make([]string, 0, len(formats))
	for _, item := range formats {
		if item == value {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
