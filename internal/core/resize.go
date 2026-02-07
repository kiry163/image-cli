package core

import (
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/pkg/apperror"
)

type ResizeOptions struct {
	Width              string
	Height             string
	Fit                string
	WithoutEnlargement bool
	KeepRatio          bool
	Conflict           string
}

func Resize(inputPath, outputArg string, opts ResizeOptions) (string, error) {
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
		DesiredFormat: "",
		InputFormat:   inputFormat,
		Conflict:      opts.Conflict,
		Overwrite:     false,
	})
	if err != nil {
		return "", err
	}
	imageSize, err := bimg.Size(buf)
	if err != nil {
		return "", apperror.InvalidInput("无法读取图像尺寸", err)
	}
	width, err := parseDimension(opts.Width, imageSize.Width)
	if err != nil {
		return "", err
	}
	height, err := parseDimension(opts.Height, imageSize.Height)
	if err != nil {
		return "", err
	}
	if width == 0 && height == 0 {
		return "", apperror.InvalidArgument("必须指定宽度或高度", nil)
	}
	outType, err := ImageTypeFromFormat(outFormat)
	if err != nil {
		return "", err
	}
	if !bimg.IsTypeSupportedSave(outType) {
		return "", apperror.UnsupportedFormat("当前环境不支持输出格式", nil)
	}
	options := bimg.Options{
		Width:   width,
		Height:  height,
		Type:    outType,
		Enlarge: !opts.WithoutEnlargement,
	}
	if err := applyFit(&options, opts.Fit); err != nil {
		return "", err
	}
	if !opts.KeepRatio {
		options.Force = true
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

func parseDimension(value string, base int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	if strings.HasSuffix(value, "%") {
		if base <= 0 {
			return 0, apperror.InvalidArgument("无法使用百分比尺寸", nil)
		}
		number := strings.TrimSuffix(value, "%")
		percent, err := strconv.ParseFloat(number, 64)
		if err != nil || percent <= 0 {
			return 0, apperror.InvalidArgument("百分比尺寸无效", err)
		}
		return int(math.Round(float64(base) * percent / 100.0)), nil
	}
	val, err := strconv.Atoi(value)
	if err != nil || val <= 0 {
		return 0, apperror.InvalidArgument("尺寸必须为正数", err)
	}
	return val, nil
}

func applyFit(options *bimg.Options, fit string) error {
	fit = strings.TrimSpace(strings.ToLower(fit))
	if fit == "" {
		return nil
	}
	switch fit {
	case "cover":
		options.Crop = true
	case "contain":
		options.Embed = true
	case "fill":
		options.Force = true
	case "inside":
		options.Embed = true
	case "outside":
		options.Crop = true
	default:
		return apperror.InvalidArgument("fit 参数无效", nil)
	}
	return nil
}
