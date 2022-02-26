package main

import (
	"fmt"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error getting current working directory: %s\n", os.Args[0], err)
		os.Exit(1)
	}
	env := os.Environ()
	args, err := newArguments(os.Args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error parsing cli arguments: %s\n", os.Args[0], err)
		os.Exit(1)
	}
	_, exitStatus := ltf(cwd, args, env)
	os.Exit(exitStatus)
}
