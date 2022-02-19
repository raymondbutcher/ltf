package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func command(cwd string, args []string, env []string) (*exec.Cmd, error) {
	// Builds and returns a command to run.
	// This also prints messages to stderr and stdout.

	subcommand, helpFlag, versionFlag := parseArgs(args)

	var cmd *exec.Cmd
	var err error

	if helpFlag {
		fmt.Println("LTF is a transparent wrapper for Terraform, so usage is no different from")
		fmt.Println("Terraform, which is detailed below. LTF checks the directory tree for")
		fmt.Println("configuration files, variables files, and backend files, and then")
		fmt.Println("alters the Terraform command and environment to use them.")
		fmt.Println("")
		cmd = terraformCommand(args)
	} else if subcommand == "" || subcommand == "fmt" || subcommand == "version" || versionFlag {
		cmd = terraformCommand(args)
	} else {
		cmd, err = wrapperCommand(cwd, args, env)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "LTF: %s\n\n", strings.Join(cmd.Args, " "))
	}

	return cmd, nil
}

func parseArgs(args []string) (subcommand string, help bool, version bool) {
	// Returns the important details of the CLI arguments.

	for _, arg := range args[1:] {
		if subcommand == "" && len(arg) > 0 && arg[0:1] != "-" {
			subcommand = arg
			break
		} else if arg == "-help" {
			help = true
		} else if arg == "-version" {
			version = true
		}
	}
	return subcommand, help, version
}

func terraformCommand(args []string) *exec.Cmd {
	// Returns a command that just runs Terraform with no changes.
	cmd := exec.Command("terraform", args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func main() {
	// Get the calling environment.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "LTF: error getting current working directory: %s\n", err)
		os.Exit(1)
	}
	args := os.Args
	env := os.Environ()

	// Build the command.
	cmd, err := command(cwd, args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LTF: %s\n", err)
		os.Exit(1)
	}

	// Run the command.
	if err := cmd.Run(); err != nil {
		if exitErr, isExitError := err.(*exec.ExitError); isExitError {
			os.Exit(exitErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Error running Terraform: %s\n", err)
			os.Exit(1)
		}
	}
}
