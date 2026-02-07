package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/pkg/apperror"
)

type ConvertOptions struct {
	Format    string
	Quality   int
	Overwrite bool
	Conflict  string
	ICOSizes  []int
}

func Convert(inputPath, outputArg string, opts ConvertOptions) (string, error) {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return "", apperror.InvalidInput("文件不存在或无法读取", err)
	}
	inputType := bimg.DetermineImageType(buf)
	if inputType == bimg.UNKNOWN {
		return "", apperror.UnsupportedFormat("无法识别输入格式", nil)
	}
	inputFormat := FormatFromImageType(inputType)
	outPath, outFormat, err := ResolveOutput(OutputSpec{
		InputPath:     inputPath,
		OutputArg:     outputArg,
		DesiredFormat: opts.Format,
		InputFormat:   inputFormat,
		Conflict:      opts.Conflict,
		Overwrite:     opts.Overwrite,
	})
	if err != nil {
		return "", err
	}
	if outFormat == "ico" {
		return convertToICO(buf, outPath, opts.ICOSizes)
	}
	outType, err := ImageTypeFromFormat(outFormat)
	if err != nil {
		return "", err
	}
	if !bimg.IsTypeSupportedSave(outType) {
		return "", apperror.UnsupportedFormat("当前环境不支持输出格式", nil)
	}
	options := bimg.Options{Type: outType}
	if opts.Quality > 0 {
		options.Quality = opts.Quality
	}
	newImage, err := bimg.NewImage(buf).Process(options)
	if err != nil {
		return "", apperror.InvalidInput("图像处理失败", err)
	}
	if err := os.WriteFile(outPath, newImage, 0o644); err != nil {
		return "", apperror.ConfigError("无法写入输出文件", err)
	}
	return outPath, nil
}

func convertToICO(buf []byte, outPath string, sizes []int) (string, error) {
	cmdPath, ok := ImageMagickCommand()
	if !ok {
		return "", apperror.UnsupportedFormat("ICO 需要安装 ImageMagick", nil)
	}
	if err := ensureParentDir(outPath); err != nil {
		return "", err
	}
	if len(sizes) == 0 {
		sizes = []int{256, 128, 64, 48, 32, 16}
	}
	tempDir, err := os.MkdirTemp("", "image-cli-ico-")
	if err != nil {
		return "", apperror.ConfigError("无法创建临时目录", err)
	}
	defer os.RemoveAll(tempDir)
	inputPath := filepath.Join(tempDir, "input.png")
	options := bimg.Options{Type: bimg.PNG}
	newImage, err := bimg.NewImage(buf).Process(options)
	if err != nil {
		return "", apperror.InvalidInput("图像处理失败", err)
	}
	if err := os.WriteFile(inputPath, newImage, 0o644); err != nil {
		return "", apperror.ConfigError("无法写入临时文件", err)
	}
	args := []string{
		inputPath,
		"-define",
		"icon:auto-resize=" + formatICOSizes(sizes),
		outPath,
	}
	cmd := exec.Command(cmdPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", apperror.ConfigError("ICO 转换失败: "+strings.TrimSpace(string(output)), err)
	}
	return outPath, nil
}

func formatICOSizes(sizes []int) string {
	parts := make([]string, 0, len(sizes))
	for _, size := range sizes {
		parts = append(parts, strconv.Itoa(size))
	}
	return strings.Join(parts, ",")
}
