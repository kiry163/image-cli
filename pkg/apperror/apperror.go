package apperror

import "fmt"

type AppError struct {
	Code    string
	Message string
	Detail  string
	Err     error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message, detail string, err error) *AppError {
	return &AppError{Code: code, Message: message, Detail: detail, Err: err}
}

func NotImplemented(message string) *AppError {
	return New("E900", message, "功能尚未实现", nil)
}

func AINotImplemented() *AppError {
	return New("E104", "AI 功能未实现", "当前版本不包含 AI 能力", nil)
}

func InvalidInput(detail string, err error) *AppError {
	return New("E001", "无效的输入文件", detail, err)
}

func UnsupportedFormat(detail string, err error) *AppError {
	return New("E002", "不支持的格式", detail, err)
}

func InvalidArgument(detail string, err error) *AppError {
	return New("E007", "参数错误", detail, err)
}

func OutputExists(detail string) *AppError {
	return New("E006", "输出文件已存在", detail, nil)
}

func BatchFailed(detail string) *AppError {
	return New("E008", "批量处理部分失败", detail, nil)
}

func ConfigError(detail string, err error) *AppError {
	return New("E005", "配置错误", detail, err)
}
