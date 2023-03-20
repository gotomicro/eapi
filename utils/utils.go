package utils

import (
	"os"
	"strings"
)

func Debug() bool {
	return os.Getenv("DEBUG") == "true" || strings.HasSuffix(os.Args[0], ".test")
}
