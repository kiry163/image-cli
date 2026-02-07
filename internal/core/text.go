package core

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiry163/image-cli/pkg/apperror"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func renderTextWatermark(text string, fontSize int, fontName string, fontFile string, colorValue string, opacity float64, strokeColorValue string, strokeWidth int, backgroundValue string, strokeMode string) ([]byte, error) {
	if text == "" {
		return nil, apperror.InvalidArgument("文字水印内容为空", nil)
	}
	if fontSize <= 0 {
		fontSize = 24
	}
	if opacity <= 0 || opacity > 1 {
		opacity = 0.5
	}
	if strokeMode == "" {
		strokeMode = "circle"
	}
	fontData, err := loadFontData(fontFile, fontName)
	if err != nil {
		return nil, err
	}
	face, err := newFontFace(fontData, fontSize)
	if err != nil {
		return nil, err
	}
	defer face.Close()

	bounds, advance := font.BoundString(face, text)
	width := (bounds.Max.X - bounds.Min.X).Ceil()
	height := (bounds.Max.Y - bounds.Min.Y).Ceil()
	padding := strokeWidth + 4
	imgW := width + padding*2
	imgH := height + padding*2
	if imgW < 1 {
		imgW = 1
	}
	if imgH < 1 {
		imgH = 1
	}
	_ = advance

	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	background, hasBackground := parseColor(backgroundValue)
	if hasBackground {
		background.A = applyOpacity(background.A, opacity)
		draw.Draw(img, img.Bounds(), &image.Uniform{C: background}, image.Point{}, draw.Src)
	}

	fillColor, ok := parseColor(colorValue)
	if !ok {
		fillColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	fillColor.A = applyOpacity(fillColor.A, opacity)
	strokeColor, ok := parseColor(strokeColorValue)
	if !ok {
		strokeColor = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}
	strokeColor.A = applyOpacity(strokeColor.A, opacity)

	baseX := padding - bounds.Min.X.Ceil()
	baseY := padding - bounds.Min.Y.Ceil()

	if strokeWidth > 0 {
		for _, offset := range strokeOffsets(strokeWidth, strokeMode) {
			drawText(img, face, baseX+offset.X, baseY+offset.Y, text, strokeColor)
		}
	}
	drawText(img, face, baseX, baseY, text, fillColor)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, apperror.ConfigError("文字水印编码失败", err)
	}
	return buf.Bytes(), nil
}

func loadFontData(fontFile string, fontName string) ([]byte, error) {
	if fontFile != "" {
		data, err := os.ReadFile(fontFile)
		if err != nil {
			return nil, apperror.ConfigError("无法读取字体文件", err)
		}
		return data, nil
	}
	if fontName != "" {
		if isFontFilePath(fontName) || strings.Contains(fontName, string(os.PathSeparator)) {
			data, err := os.ReadFile(fontName)
			if err != nil {
				return nil, apperror.ConfigError("无法读取字体文件", err)
			}
			return data, nil
		}
		if strings.EqualFold(fontName, defaultFontName) {
			return defaultFontData, nil
		}
		return nil, apperror.ConfigError("字体名称无法解析，请使用 --font-file 指定字体文件", nil)
	}
	if len(defaultFontData) == 0 {
		return nil, apperror.ConfigError("内置字体不可用", nil)
	}
	return defaultFontData, nil
}

func newFontFace(fontData []byte, fontSize int) (font.Face, error) {
	f, err := opentype.Parse(fontData)
	if err != nil {
		return nil, apperror.ConfigError("字体解析失败", err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    float64(fontSize),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, apperror.ConfigError("字体加载失败", err)
	}
	return face, nil
}

func drawText(img *image.RGBA, face font.Face, x, y int, text string, c color.NRGBA) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

func strokeOffsets(width int, mode string) []image.Point {
	offsets := []image.Point{}
	if width <= 0 {
		return offsets
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "8dir" || mode == "8" {
		for i := 1; i <= width; i++ {
			offsets = append(offsets,
				image.Point{X: i, Y: 0},
				image.Point{X: -i, Y: 0},
				image.Point{X: 0, Y: i},
				image.Point{X: 0, Y: -i},
				image.Point{X: i, Y: i},
				image.Point{X: -i, Y: -i},
				image.Point{X: i, Y: -i},
				image.Point{X: -i, Y: i},
			)
		}
		return offsets
	}
	for dy := -width; dy <= width; dy++ {
		for dx := -width; dx <= width; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if dx*dx+dy*dy <= width*width {
				offsets = append(offsets, image.Point{X: dx, Y: dy})
			}
		}
	}
	return offsets
}

func parseColor(value string) (color.NRGBA, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == "none" || value == "transparent" {
		return color.NRGBA{}, false
	}
	if strings.HasPrefix(value, "#") {
		hex := strings.TrimPrefix(value, "#")
		if len(hex) == 6 || len(hex) == 8 {
			r, _ := parseHexByte(hex[0:2])
			g, _ := parseHexByte(hex[2:4])
			b, _ := parseHexByte(hex[4:6])
			a := byte(255)
			if len(hex) == 8 {
				a, _ = parseHexByte(hex[6:8])
			}
			return color.NRGBA{R: r, G: g, B: b, A: a}, true
		}
	}
	if strings.HasPrefix(value, "rgba(") && strings.HasSuffix(value, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(value, "rgba("), ")")
		parts := splitComma(inner)
		if len(parts) == 4 {
			r, _ := parseIntByte(parts[0])
			g, _ := parseIntByte(parts[1])
			b, _ := parseIntByte(parts[2])
			alpha := parseAlpha(parts[3])
			return color.NRGBA{R: r, G: g, B: b, A: alpha}, true
		}
	}
	if strings.HasPrefix(value, "rgb(") && strings.HasSuffix(value, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(value, "rgb("), ")")
		parts := splitComma(inner)
		if len(parts) == 3 {
			r, _ := parseIntByte(parts[0])
			g, _ := parseIntByte(parts[1])
			b, _ := parseIntByte(parts[2])
			return color.NRGBA{R: r, G: g, B: b, A: 255}, true
		}
	}
	if c, ok := namedColors[value]; ok {
		return c, true
	}
	return color.NRGBA{}, false
}

func parseHexByte(s string) (byte, error) {
	var v byte
	_, err := fmt.Sscanf(s, "%02x", &v)
	return v, err
}

func parseIntByte(s string) (byte, error) {
	var v int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &v)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return byte(v), nil
}

func parseAlpha(s string) byte {
	value := strings.TrimSpace(s)
	if strings.Contains(value, ".") {
		var f float64
		_, err := fmt.Sscanf(value, "%f", &f)
		if err != nil {
			return 255
		}
		if f < 0 {
			f = 0
		}
		if f > 1 {
			f = 1
		}
		return byte(math.Round(f * 255))
	}
	val, err := parseIntByte(value)
	if err != nil {
		return 255
	}
	return val
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func applyOpacity(alpha byte, opacity float64) byte {
	if opacity <= 0 {
		return 0
	}
	if opacity >= 1 {
		return alpha
	}
	return byte(math.Round(float64(alpha) * opacity))
}

var namedColors = map[string]color.NRGBA{
	"black":   {R: 0, G: 0, B: 0, A: 255},
	"white":   {R: 255, G: 255, B: 255, A: 255},
	"red":     {R: 255, G: 0, B: 0, A: 255},
	"green":   {R: 0, G: 128, B: 0, A: 255},
	"blue":    {R: 0, G: 0, B: 255, A: 255},
	"yellow":  {R: 255, G: 255, B: 0, A: 255},
	"cyan":    {R: 0, G: 255, B: 255, A: 255},
	"magenta": {R: 255, G: 0, B: 255, A: 255},
}

func isFontFilePath(path string) bool {
	if path == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".ttf" || ext == ".otf" || ext == ".ttc"
}
