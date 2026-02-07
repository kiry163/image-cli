package core

import (
	"os"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/pkg/apperror"
)

type RotateOptions struct {
	Degrees  int
	Flip     bool
	Flop     bool
	Conflict string
}

func Rotate(inputPath, outputArg string, opts RotateOptions) (string, error) {
	if opts.Degrees == 0 && !opts.Flip && !opts.Flop {
		return "", apperror.InvalidArgument("必须指定旋转角度或翻转", nil)
	}
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
	outType, err := ImageTypeFromFormat(outFormat)
	if err != nil {
		return "", err
	}
	if !bimg.IsTypeSupportedSave(outType) {
		return "", apperror.UnsupportedFormat("当前环境不支持输出格式", nil)
	}
	angle, err := parseAngle(opts.Degrees)
	if err != nil {
		return "", err
	}
	options := bimg.Options{
		Type:   outType,
		Rotate: angle,
		Flip:   opts.Flip,
		Flop:   opts.Flop,
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

func parseAngle(degrees int) (bimg.Angle, error) {
	if degrees == 0 {
		return bimg.D0, nil
	}
	switch degrees {
	case 90, -270:
		return bimg.D90, nil
	case 180, -180:
		return bimg.D180, nil
	case 270, -90:
		return bimg.D270, nil
	default:
		return bimg.D0, apperror.InvalidArgument("旋转角度仅支持 90/180/270/-90", nil)
	}
}
