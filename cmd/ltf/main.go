package main

import (
	"fmt"
	"os"

	"github.com/raymondbutcher/ltf"
	"github.com/raymondbutcher/ltf/internal/arguments"
	internal "github.com/raymondbutcher/ltf/internal/ltf" // TODO: refactor this package away
)

// version is updated with -ldflags in the pipeline.
var version = "development"

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error getting current working directory: %s\n", os.Args[0], err)
		os.Exit(1)
	}

	env := ltf.NewEnviron(os.Environ()...)
	args, err := arguments.New(os.Args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error parsing cli arguments: %s\n", os.Args[0], err)
		os.Exit(1)
	}

	_, exitStatus, err := internal.Run(cwd, args, env, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", args.Bin, err)
	}

	os.Exit(exitStatus)
}
