package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiry163/image-cli/pkg/apperror"
)

type OutputSpec struct {
	InputPath     string
	OutputArg     string
	DesiredFormat string
	InputFormat   string
	Conflict      string
	Overwrite     bool
}

func ResolveOutput(spec OutputSpec) (string, string, error) {
	if strings.TrimSpace(spec.OutputArg) == "" {
		return "", "", apperror.InvalidArgument("输出路径不能为空", nil)
	}
	isDir, err := outputIsDir(spec.OutputArg)
	if err != nil {
		return "", "", apperror.ConfigError("无法读取输出路径", err)
	}
	format := NormalizeFormat(spec.DesiredFormat)
	inputFormat := NormalizeFormat(spec.InputFormat)
	if isDir {
		if format == "" {
			format = inputFormat
		}
		if format == "" {
			return "", "", apperror.UnsupportedFormat("无法推断输出格式", nil)
		}
		base := filepath.Base(spec.InputPath)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		outPath := filepath.Join(spec.OutputArg, fmt.Sprintf("%s.%s", name, format))
		outPath, err = applyConflict(outPath, spec.Conflict, spec.Overwrite)
		if err != nil {
			return "", "", err
		}
		if err := ensureParentDir(outPath); err != nil {
			return "", "", err
		}
		return outPath, format, nil
	}
	outPath := spec.OutputArg
	ext := strings.TrimPrefix(filepath.Ext(outPath), ".")
	if ext != "" {
		if format == "" {
			format = ext
		}
	} else {
		if format == "" {
			format = inputFormat
		}
		if format == "" {
			return "", "", apperror.UnsupportedFormat("无法推断输出格式", nil)
		}
		outPath = outPath + "." + format
	}
	outPath, err = applyConflict(outPath, spec.Conflict, spec.Overwrite)
	if err != nil {
		return "", "", err
	}
	if err := ensureParentDir(outPath); err != nil {
		return "", "", err
	}
	return outPath, format, nil
}

func outputIsDir(path string) (bool, error) {
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return true, nil
	}
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ensureParentDir(path string) error {
	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return apperror.ConfigError("无法创建输出目录", err)
	}
	return nil
}

func applyConflict(path string, conflict string, overwrite bool) (string, error) {
	if overwrite || conflict == "overwrite" {
		return path, nil
	}
	if _, err := os.Stat(path); err == nil {
		if conflict == "rename" {
			return renamePath(path)
		}
		return "", apperror.OutputExists("输出文件已存在: " + path)
	} else if !os.IsNotExist(err) {
		return "", apperror.ConfigError("无法访问输出路径", err)
	}
	return path, nil
}

func renamePath(path string) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", apperror.OutputExists("无法生成可用的输出文件名")
}
