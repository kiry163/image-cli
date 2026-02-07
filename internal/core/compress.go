package core

import (
	"os"
	"strconv"
	"strings"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/pkg/apperror"
)

type CompressOptions struct {
	Quality        int
	MaxSizeBytes   int64
	Aggressive     bool
	Conflict       string
	DefaultQuality int
}

func Compress(inputPath, outputArg string, opts CompressOptions) (string, error) {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return "", apperror.InvalidInput("文件不存在或无法读取", err)
	}
	inputType := bimg.DetermineImageType(buf)
	if inputType == bimg.UNKNOWN {
		return "", apperror.UnsupportedFormat("无法识别输入格式", nil)
	}
	inputFormat := FormatFromImageType(inputType)
	if opts.Quality <= 0 {
		opts.Quality = opts.DefaultQuality
		if opts.Quality <= 0 {
			opts.Quality = 85
		}
	}
	outPath, outFormat, err := ResolveOutput(OutputSpec{
		InputPath:     inputPath,
		OutputArg:     outputArg,
		DesiredFormat: "",
		InputFormat:   inputFormat,
		Conflict:      opts.Conflict,
		Overwrite:     false,
	})
	if err != nil {
		return "", err
	}
	outType, err := ImageTypeFromFormat(outFormat)
	if err != nil {
		return "", err
	}
	if !bimg.IsTypeSupportedSave(outType) {
		return "", apperror.UnsupportedFormat("当前环境不支持输出格式", nil)
	}
	quality := opts.Quality
	newImage, err := processWithQuality(buf, outType, quality)
	if err != nil {
		return "", err
	}
	if opts.MaxSizeBytes > 0 {
		step := 2
		minQuality := 10
		if opts.Aggressive {
			step = 5
			minQuality = 5
		}
		for int64(len(newImage)) > opts.MaxSizeBytes && quality > minQuality {
			quality -= step
			newImage, err = processWithQuality(buf, outType, quality)
			if err != nil {
				return "", err
			}
		}
	}
	if err := os.WriteFile(outPath, newImage, 0o644); err != nil {
		return "", apperror.ConfigError("无法写入输出文件", err)
	}
	return outPath, nil
}

func ParseSizeBytes(value string) (int64, error) {
	value = strings.TrimSpace(strings.ToUpper(value))
	if value == "" {
		return 0, nil
	}
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(value, "KB"):
		multiplier = 1024
		value = strings.TrimSuffix(value, "KB")
	case strings.HasSuffix(value, "MB"):
		multiplier = 1024 * 1024
		value = strings.TrimSuffix(value, "MB")
	case strings.HasSuffix(value, "GB"):
		multiplier = 1024 * 1024 * 1024
		value = strings.TrimSuffix(value, "GB")
	case strings.HasSuffix(value, "B"):
		value = strings.TrimSuffix(value, "B")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, apperror.InvalidArgument("最大文件大小无效", nil)
	}
	val, err := strconv.ParseFloat(value, 64)
	if err != nil || val <= 0 {
		return 0, apperror.InvalidArgument("最大文件大小无效", err)
	}
	return int64(val * float64(multiplier)), nil
}

func processWithQuality(buf []byte, outType bimg.ImageType, quality int) ([]byte, error) {
	options := bimg.Options{Type: outType}
	if quality > 0 {
		options.Quality = quality
	}
	newImage, err := bimg.NewImage(buf).Process(options)
	if err != nil {
		return nil, apperror.InvalidInput("图像处理失败", err)
	}
	return newImage, nil
}
