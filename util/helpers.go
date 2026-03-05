package util

import (
	"fmt"
	"os"
)

// FatalIfError prints the error and exits with code 1 if non-nil.
// Standard CLI error handler.
func FatalIfError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
