package util

import (
	"fmt"
	"os"
)

func ExitIfError(err error) {
	ExitfIfError(err, "an unexpected error occurred")
}

func ExitfIfError(err error, message string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("%s: %w", message, err))
		os.Exit(1)
	}
}

func Exitf(messageFormat string, args ...interface{}) {
	ExitIfError(fmt.Errorf(messageFormat, args...))
}
