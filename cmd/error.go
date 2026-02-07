package cmd

import (
	"fmt"
	"io"

	"github.com/kiry163/image-cli/pkg/apperror"
)

func WriteError(w io.Writer, err error) {
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		fmt.Fprintln(w, err.Error())
		return
	}
	fmt.Fprintf(w, "Error [%s]: %s\n", appErr.Code, appErr.Message)
	if appErr.Detail != "" {
		fmt.Fprintf(w, "  → %s\n", appErr.Detail)
	}
}
