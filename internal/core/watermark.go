package core

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
	Color       string
	StrokeColor string
	StrokeWidth int
	Background  string
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
		return "", apperror.InvalidInput("图像处理失败", err)
	}
	if err := os.WriteFile(outPath, newImage, 0o644); err != nil {
		return "", apperror.ConfigError("无法写入输出文件", err)
	}
	return outPath, nil
}

func buildWatermarkBuffer(baseBuf []byte, opts WatermarkOptions) ([]byte, error) {
	if opts.Text != "" {
		return renderTextWatermark(opts.Text, opts.FontSize, opts.Font, opts.Color, opts.Opacity, opts.StrokeColor, opts.StrokeWidth, opts.Background)
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

func renderTextWatermark(text string, fontSize int, font, color string, opacity float64, strokeColor string, strokeWidth int, background string) ([]byte, error) {
	cmdPath, ok := ImageMagickCommand()
	if !ok {
		return nil, apperror.UnsupportedFormat("文本水印需要安装 ImageMagick", nil)
	}
	if fontSize <= 0 {
		fontSize = 24
	}
	if color == "" {
		color = "white"
	}
	if background == "" {
		background = "none"
	}
	if opacity <= 0 || opacity > 1 {
		opacity = 0.5
	}
	fillColor := applyOpacity(color, opacity)
	stroke := ""
	if strokeWidth > 0 {
		if strokeColor == "" {
			strokeColor = "black"
		}
		stroke = applyOpacity(strokeColor, opacity)
	}
	args := []string{
		"-background",
		background,
		"-fill",
		fillColor,
		"-pointsize",
		fmt.Sprintf("%d", fontSize),
		"label:" + text,
		"png:-",
	}
	if strokeWidth > 0 {
		args = append([]string{
			"-stroke",
			stroke,
			"-strokewidth",
			fmt.Sprintf("%d", strokeWidth),
		}, args...)
	}
	if font != "" {
		args = append([]string{"-font", font}, args...)
	}
	cmd := exec.Command(cmdPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = "文本水印生成失败"
		}
		return nil, apperror.ConfigError(errMsg, err)
	}
	return output, nil
}

func applyOpacity(color string, opacity float64) string {
	color = strings.TrimSpace(color)
	if color == "" || color == "none" || color == "transparent" {
		return color
	}
	if opacity <= 0 || opacity >= 1 {
		return color
	}
	if strings.Contains(color, "@") || strings.HasPrefix(color, "rgba(") || strings.HasPrefix(color, "hsla(") {
		return color
	}
	return fmt.Sprintf("%s@%.3f", color, opacity)
}
