package util

import (
	"fmt"
	"os"
)

func FatalIfError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}