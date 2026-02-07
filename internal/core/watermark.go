package core

import (
	"os"
	"strings"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/pkg/apperror"
)

type WatermarkOptions struct {
	LogoPath    string
	Text        string
	Opacity     float64
	Scale       float64
	Gravity     string
	OffsetX     int
	OffsetY     int
	FontSize    int
	Font        string
	FontFile    string
	Color       string
	StrokeColor string
	StrokeWidth int
	Background  string
	StrokeMode  string
	Conflict    string
}

func Watermark(inputPath, outputArg string, opts WatermarkOptions) (string, error) {
	if opts.Text != "" && opts.LogoPath != "" {
		return "", apperror.InvalidArgument("文本水印与图片水印不可同时使用", nil)
	}
	if opts.Text == "" && opts.LogoPath == "" {
		return "", apperror.InvalidArgument("必须提供水印图片或文本", nil)
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
	if opts.Opacity <= 0 || opts.Opacity > 1 {
		return "", apperror.InvalidArgument("不透明度必须在 0-1 之间", nil)
	}
	watermarkBuf, err := buildWatermarkBuffer(buf, opts)
	if err != nil {
		return "", err
	}
	baseSize, err := bimg.Size(buf)
	if err != nil {
		return "", apperror.InvalidInput("无法读取图像尺寸", err)
	}
	wmSize, err := bimg.Size(watermarkBuf)
	if err != nil {
		return "", apperror.InvalidInput("无法读取水印尺寸", err)
	}
	watermarkBuf, wmSize, err = scaleWatermarkIfNeeded(watermarkBuf, wmSize, baseSize)
	if err != nil {
		return "", err
	}
	left, top, err := gravityPosition(baseSize.Width, baseSize.Height, wmSize.Width, wmSize.Height, opts.Gravity, opts.OffsetX, opts.OffsetY)
	if err != nil {
		return "", err
	}
	options := bimg.Options{
		Type: outType,
		WatermarkImage: bimg.WatermarkImage{
			Left:    left,
			Top:     top,
			Buf:     watermarkBuf,
			Opacity: float32(opts.Opacity),
		},
	}
	newImage, err := bimg.NewImage(buf).Process(options)
	if err != nil {
		if opts.Text != "" {
			return "", apperror.InvalidInput("文字水印处理失败，建议指定字体或缩小字号", err)
		}
		return "", apperror.InvalidInput("图像处理失败", err)
	}
	if err := os.WriteFile(outPath, newImage, 0o644); err != nil {
		return "", apperror.ConfigError("无法写入输出文件", err)
	}
	return outPath, nil
}

func scaleWatermarkIfNeeded(buf []byte, wmSize bimg.ImageSize, baseSize bimg.ImageSize) ([]byte, bimg.ImageSize, error) {
	if wmSize.Width <= 0 || wmSize.Height <= 0 || baseSize.Width <= 0 || baseSize.Height <= 0 {
		return buf, wmSize, nil
	}
	maxW := int(float64(baseSize.Width) * 0.9)
	maxH := int(float64(baseSize.Height) * 0.9)
	if wmSize.Width <= maxW && wmSize.Height <= maxH {
		return buf, wmSize, nil
	}
	scaleW := float64(maxW) / float64(wmSize.Width)
	scaleH := float64(maxH) / float64(wmSize.Height)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}
	newW := int(float64(wmSize.Width) * scale)
	newH := int(float64(wmSize.Height) * scale)
	if newW <= 0 {
		newW = 1
	}
	if newH <= 0 {
		newH = 1
	}
	resized, err := bimg.NewImage(buf).Process(bimg.Options{Width: newW, Height: newH})
	if err != nil {
		return nil, wmSize, apperror.InvalidInput("水印缩放失败", err)
	}
	newSize := bimg.ImageSize{Width: newW, Height: newH}
	return resized, newSize, nil
}

func buildWatermarkBuffer(baseBuf []byte, opts WatermarkOptions) ([]byte, error) {
	if opts.Text != "" {
		return renderTextWatermark(opts.Text, opts.FontSize, opts.Font, opts.FontFile, opts.Color, opts.Opacity, opts.StrokeColor, opts.StrokeWidth, opts.Background, opts.StrokeMode)
	}
	logoBuf, err := os.ReadFile(opts.LogoPath)
	if err != nil {
		return nil, apperror.InvalidInput("无法读取水印图片", err)
	}
	if opts.Scale <= 0 || opts.Scale > 1 {
		return nil, apperror.InvalidArgument("水印缩放比例必须在 0-1 之间", nil)
	}
	baseSize, err := bimg.Size(baseBuf)
	if err != nil {
		return nil, apperror.InvalidInput("无法读取图像尺寸", err)
	}
	wmSize, err := bimg.Size(logoBuf)
	if err != nil {
		return nil, apperror.InvalidInput("无法读取水印尺寸", err)
	}
	shortSide := baseSize.Width
	if baseSize.Height < shortSide {
		shortSide = baseSize.Height
	}
	target := int(float64(shortSide) * opts.Scale)
	if target <= 0 {
		return nil, apperror.InvalidArgument("水印缩放比例过小", nil)
	}
	var resizeOptions bimg.Options
	if wmSize.Width >= wmSize.Height {
		resizeOptions = bimg.Options{Width: target}
	} else {
		resizeOptions = bimg.Options{Height: target}
	}
	resized, err := bimg.NewImage(logoBuf).Process(resizeOptions)
	if err != nil {
		return nil, apperror.InvalidInput("水印缩放失败", err)
	}
	return resized, nil
}

func gravityPosition(baseW, baseH, wmW, wmH int, gravity string, offsetX, offsetY int) (int, int, error) {
	gravity = strings.ToLower(strings.TrimSpace(gravity))
	if gravity == "" {
		gravity = "southeast"
	}
	var left int
	var top int
	switch gravity {
	case "northwest":
		left, top = 0, 0
	case "north":
		left, top = (baseW-wmW)/2, 0
	case "northeast":
		left, top = baseW-wmW, 0
	case "west":
		left, top = 0, (baseH-wmH)/2
	case "center":
		left, top = (baseW-wmW)/2, (baseH-wmH)/2
	case "east":
		left, top = baseW-wmW, (baseH-wmH)/2
	case "southwest":
		left, top = 0, baseH-wmH
	case "south":
		left, top = (baseW-wmW)/2, baseH-wmH
	case "southeast":
		left, top = baseW-wmW, baseH-wmH
	default:
		return 0, 0, apperror.InvalidArgument("gravity 参数无效", nil)
	}
	left += offsetX
	top += offsetY
	if left < 0 {
		left = 0
	}
	if top < 0 {
		top = 0
	}
	if left > baseW-wmW {
		left = baseW - wmW
	}
	if top > baseH-wmH {
		top = baseH - wmH
	}
	return left, top, nil
}
