package core

import (
	"strings"

	"github.com/h2non/bimg"
	"github.com/kiry163/image-cli/pkg/apperror"
)

func NormalizeFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	format = strings.TrimPrefix(format, ".")
	if format == "jpeg" {
		return "jpg"
	}
	if format == "tif" {
		return "tiff"
	}
	return format
}

func ImageTypeFromFormat(format string) (bimg.ImageType, error) {
	format = NormalizeFormat(format)
	switch format {
	case "jpg":
		return bimg.JPEG, nil
	case "png":
		return bimg.PNG, nil
	case "webp":
		return bimg.WEBP, nil
	case "gif":
		return bimg.GIF, nil
	case "tiff":
		return bimg.TIFF, nil
	case "pdf":
		return bimg.PDF, nil
	case "heif", "heic":
		return bimg.HEIF, nil
	case "avif":
		return bimg.AVIF, nil
	case "svg":
		return bimg.SVG, nil
	default:
		return bimg.UNKNOWN, apperror.UnsupportedFormat("未知输出格式", nil)
	}
}

func FormatFromImageType(imageType bimg.ImageType) string {
	name := bimg.ImageTypeName(imageType)
	if name == "jpeg" {
		return "jpg"
	}
	if name == "tif" {
		return "tiff"
	}
	return name
}
