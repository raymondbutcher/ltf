package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func command(cwd string, args []string, env []string, config *Config) (*exec.Cmd, error) {
	// Builds and returns a command to run.

	subcommand, helpFlag, versionFlag := parseArgs(args)

	var cmd *exec.Cmd
	var err error

	if helpFlag || versionFlag || subcommand == "" || subcommand == "fmt" || subcommand == "version" {
		cmd = terraformCommand(args)
	} else {
		cmd, err = wrapperCommand(cwd, args, env)
		if err != nil {
			return nil, err
		}
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
		fmt.Fprintf(os.Stderr, "[LTF] error getting current working directory: %s\n", err)
		os.Exit(1)
	}
	args := os.Args
	env := os.Environ()
	_, helpFlag, _ := parseArgs(args)

	// Print environment variables for hooks.
	if args[1] == "-ltf-env-to-json" {
		envJsonBytes, err := json.Marshal(os.Environ())
		if err != nil {
			fmt.Fprintf(os.Stderr, "[LTF] error writing environment to JSON: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(string(envJsonBytes))
		os.Exit(0)
	}

	// Load the configuration YAML file.
	config, err := loadConfig(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] error loading hooks: %s\n", err)
		os.Exit(1)
	}

	// Build the command.
	cmd, err := command(cwd, args, env, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] %s\n", err)
		os.Exit(1)
	}

	// Trigger hooks.
	err = config.Trigger("before", cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] error from hook: %s\n", err)
		os.Exit(1)
	}

	// Print a help message before Terraform's help message.
	if helpFlag {
		fmt.Println("LTF is a transparent wrapper for Terraform, so usage is no different from")
		fmt.Println("Terraform, which is detailed below. LTF checks the directory tree for")
		fmt.Println("configuration files, variables files, and backend files, and then")
		fmt.Println("alters the Terraform command and environment to use them.")
		fmt.Println("")
	}

	// Run the Terraform command.
	fmt.Fprintf(os.Stderr, "[LTF] running: %s\n", strings.Join(cmd.Args, " "))
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, isExitError := err.(*exec.ExitError); isExitError {
			exitCode = exitErr.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "[LTF] Error running Terraform: %s\n", err)
			exitCode = 1
		}
	}

	// Trigger hooks.
	if exitCode == 0 {
		err = config.Trigger("after", cmd)
	} else {
		err = config.Trigger("failed", cmd)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] error from hook: %s\n", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}
