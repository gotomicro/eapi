package main

import (
	"fmt"
	"os"

	"ego-gen-api/cmd"
	_ "ego-gen-api/cmd/gen"
	_ "ego-gen-api/cmd/version"
)

func main() {
	err := cmd.RootCommand.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return

}
