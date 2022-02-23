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

func command(cwd string, args []string, env []string, settings *Settings) (*exec.Cmd, map[string]string, error) {
	// Builds and returns a command to run.

	subcommand, helpFlag, versionFlag := parseArgs(args)

	if helpFlag || versionFlag || subcommand == "" || subcommand == "fmt" || subcommand == "version" {
		// Skip the wrapper and run Terraform directly.
		cmd := exec.Command("terraform", args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		return cmd, nil, nil
	}

	return wrapperCommand(cwd, args, env)
}

func ltf(cwd string, args []string, env []string) (cmd *exec.Cmd, exitStatus int) {
	// Special mode to output environment variables after running a hook script.
	// It outputs in JSON format to avoid issues with multi-line variables.
	if len(args) > 1 && args[1] == "-ltf-env-to-json" {
		envJsonBytes, err := json.Marshal(env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[LTF] Error writing environment to JSON: %s\n", err)
			return nil, 1
		}
		fmt.Print(string(envJsonBytes))
		return nil, 0
	}

	// Find and load the optional settings file.
	settings, err := loadSettings(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error loading settings: %s\n", err)
		return nil, 1
	}

	// Build the command.
	cmd, frozen, err := command(cwd, args, env, settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error building command: %s\n", err)
		return nil, 1
	}

	// Run any "before" hooks.
	err = settings.runHooks("before", cmd, frozen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error from hook: %s\n", err)
		return nil, 1
	}

	// Print the LTF help message before Terraform's help message.
	_, helpFlag, _ := parseArgs(args)
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
		err = settings.runHooks("after", cmd, frozen)
	} else {
		err = settings.runHooks("failed", cmd, frozen)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "[LTF] Error from hook: %s\n", err)
		return nil, 1
	}

	return cmd, exitCode
}

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
