package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Get the calling environment.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "LTF: error getting current working directory: %s\n", err)
		os.Exit(1)
	}
	args := os.Args
	env := os.Environ()

	// Build the Terraform command.
	cmd, err := terraformCommand(cwd, args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LTF: %s\n", err)
		os.Exit(1)
	}

	// Run the Terraform command.
	fmt.Fprintf(os.Stderr, "LTF: %s\n", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		if exitErr, isExitError := err.(*exec.ExitError); isExitError {
			os.Exit(exitErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Error running Terraform: %s\n", err)
			os.Exit(1)
		}
	}
}
