package apperror

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	appErr, ok := err.(*AppError)
	if !ok {
		return 1
	}
	switch appErr.Code {
	case "E002", "E005", "E006", "E007", "E008":
		return 2
	case "E001", "E003", "E004":
		return 3
	case "E101", "E102", "E103", "E104":
		return 4
	default:
		return 1
	}
}
