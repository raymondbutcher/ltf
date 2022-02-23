package main

import (
	"fmt"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error getting current working directory: %s\n", err)
		os.Exit(1)
	}
	args := os.Args
	env := os.Environ()
	_, exitStatus := ltf(cwd, args, env)
	os.Exit(exitStatus)
}
