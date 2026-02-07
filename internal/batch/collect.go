package batch

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiry163/image-cli/pkg/apperror"
)

type CollectResult struct {
	BaseDir string
	Files   []string
}

func Collect(pattern string, recursive bool) (CollectResult, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return CollectResult{}, apperror.InvalidInput("输入路径不能为空", nil)
	}
	if hasGlob(pattern) {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return CollectResult{}, apperror.InvalidInput("通配符无效", err)
		}
		if len(matches) == 0 {
			return CollectResult{}, apperror.InvalidInput("未匹配到文件", nil)
		}
		files := make([]string, 0, len(matches))
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			files = append(files, match)
		}
		if len(files) == 0 {
			return CollectResult{}, apperror.InvalidInput("未匹配到文件", nil)
		}
		return CollectResult{BaseDir: filepath.Dir(pattern), Files: files}, nil
	}
	info, err := os.Stat(pattern)
	if err != nil {
		return CollectResult{}, apperror.InvalidInput("文件不存在或无法读取", err)
	}
	if !info.IsDir() {
		return CollectResult{BaseDir: filepath.Dir(pattern), Files: []string{pattern}}, nil
	}
	files, err := collectDir(pattern, recursive)
	if err != nil {
		return CollectResult{}, err
	}
	if len(files) == 0 {
		return CollectResult{}, apperror.InvalidInput("目录内没有可用文件", nil)
	}
	return CollectResult{BaseDir: pattern, Files: files}, nil
}

func collectDir(root string, recursive bool) ([]string, error) {
	files := make([]string, 0, 32)
	if recursive {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, apperror.InvalidInput("无法遍历目录", err)
		}
		return files, nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, apperror.InvalidInput("无法读取目录", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, filepath.Join(root, entry.Name()))
	}
	return files, nil
}

func hasGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}
