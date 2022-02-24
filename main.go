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
	env := os.Environ()
	args, err := newArguments(os.Args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error parsing CLI arguments: %s\n", err)
		os.Exit(1)
	}
	_, exitStatus := ltf(cwd, args, env)
	os.Exit(exitStatus)
}
