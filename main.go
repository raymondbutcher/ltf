package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const helpMessage = `[LTF] Showing help:

LTF is a transparent wrapper for Terraform; it passes all command line
arguments and environment variables through to Terraform. LTF also checks
the current directory and parent directories for various Terraform files
and alters the command line arguments and environment variables to make
Terraform use them.

LTF also executes hooks defined in the first 'ltf.yaml' file it finds
in the current directory or parent directories. This can be used to run
commands or modify the environment before and after Terraform runs.`

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
		fmt.Fprintf(os.Stderr, "[LTF] Error getting current working directory: %s\n", err)
		os.Exit(1)
	}
	args := os.Args
	env := os.Environ()
	_, helpFlag, _ := parseArgs(args)

	// Special mode to output environment variables after running a hook script.
	// It outputs in JSON format to avoid issues with multi-line variables.
	if len(args) > 1 && args[1] == "-ltf-env-to-json" {
		envJsonBytes, err := json.Marshal(os.Environ())
		if err != nil {
			fmt.Fprintf(os.Stderr, "[LTF] Error writing environment to JSON: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(string(envJsonBytes))
		os.Exit(0)
	}

	// Load the configuration YAML file.
	config, err := loadConfig(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error loading LTF configuration: %s\n", err)
		os.Exit(1)
	}

	// Build the command.
	cmd, err := command(cwd, args, env, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error building command: %s\n", err)
		os.Exit(1)
	}

	// Run any "before" hooks.
	err = config.runHooks("before", cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error from hook: %s\n", err)
		os.Exit(1)
	}

	// Print the LTF help message before Terraform's help message.
	if helpFlag {
		fmt.Println(helpMessage)
		fmt.Println("")
	}

	// Run the Terraform command.
	fmt.Fprintf(os.Stderr, "[LTF] Running: %s\n", strings.Join(cmd.Args, " "))
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, isExitError := err.(*exec.ExitError); isExitError {
			exitCode = exitErr.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "[LTF] Error running Terraform: %s\n", err)
			exitCode = 1
		}
	}

	// Run any "after" or "failed" hooks.
	if exitCode == 0 {
		err = config.runHooks("after", cmd)
	} else {
		err = config.runHooks("failed", cmd)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error from hook: %s\n", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}
