package core

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kiry163/image-cli/pkg/apperror"
)

var allowedICOSizes = map[int]struct{}{
	16:  {},
	32:  {},
	48:  {},
	64:  {},
	128: {},
	256: {},
}

func ParseICOSizes(value string) ([]int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	seen := map[int]struct{}{}
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, apperror.InvalidArgument("ICO 尺寸无效", err)
		}
		if _, ok := allowedICOSizes[val]; !ok {
			return nil, apperror.InvalidArgument("ICO 尺寸不支持", nil)
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		result = append(result, val)
	}
	if len(result) == 0 {
		return nil, apperror.InvalidArgument("ICO 尺寸无效", nil)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] > result[j]
	})
	return result, nil
}
