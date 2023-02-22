package utils

import "os"

func Debug() bool {
	return os.Getenv("DEBUG") == "true"
}
