package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/internal/core"
	"github.com/kiry163/image-cli/pkg/apperror"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <input>",
		Short: "查看图像信息",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]
			fileInfo, err := os.Stat(input)
			if err != nil {
				return apperror.InvalidInput("文件不存在或无法读取", err)
			}
			if strings.EqualFold(filepath.Ext(input), ".ico") {
				width, height, err := readICOSize(input)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "文件名: %s\n", filepath.Base(input))
				fmt.Fprintf(cmd.OutOrStdout(), "格式: ico\n")
				fmt.Fprintf(cmd.OutOrStdout(), "尺寸: %dx%d\n", width, height)
				fmt.Fprintf(cmd.OutOrStdout(), "大小: %d bytes\n", fileInfo.Size())
				return nil
			}
			buf, err := os.ReadFile(input)
			if err != nil {
				return apperror.InvalidInput("无法读取文件", err)
			}
			imageType := bimg.DetermineImageType(buf)
			if imageType == bimg.UNKNOWN {
				return apperror.UnsupportedFormat("无法识别输入格式", nil)
			}
			meta, err := bimg.Metadata(buf)
			if err != nil {
				return apperror.InvalidInput("无法解析图像", err)
			}
			format := core.FormatFromImageType(imageType)
			fmt.Fprintf(cmd.OutOrStdout(), "文件名: %s\n", filepath.Base(input))
			fmt.Fprintf(cmd.OutOrStdout(), "格式: %s\n", format)
			fmt.Fprintf(cmd.OutOrStdout(), "尺寸: %dx%d\n", meta.Size.Width, meta.Size.Height)
			fmt.Fprintf(cmd.OutOrStdout(), "大小: %d bytes\n", fileInfo.Size())
			return nil
		},
	}
	return cmd
}

func readICOSize(path string) (int, int, error) {
	cmdPath, prefix, ok := core.ImageMagickIdentifyCommand()
	if !ok {
		return 0, 0, apperror.UnsupportedFormat("ICO 需要安装 ImageMagick", nil)
	}
	args := append(prefix, "-format", "%w %h\n", path)
	output, err := exec.Command(cmdPath, args...).CombinedOutput()
	if err != nil {
		return 0, 0, apperror.InvalidInput("无法解析图像", err)
	}
	width, height, err := parseIdentifyOutput(string(output))
	if err != nil {
		return 0, 0, apperror.InvalidInput("无法解析图像", err)
	}
	return width, height, nil
}

func parseIdentifyOutput(output string) (int, int, error) {
	lines := strings.Split(output, "\n")
	maxArea := 0
	maxW := 0
	maxH := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		w, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		h, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		area := w * h
		if area > maxArea {
			maxArea = area
			maxW = w
			maxH = h
		}
	}
	if maxArea == 0 {
		return 0, 0, fmt.Errorf("no sizes")
	}
	return maxW, maxH, nil
}
