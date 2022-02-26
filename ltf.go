package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const helpMessage = `LTF is a transparent wrapper for Terraform; it passes all command line
arguments and environment variables through to Terraform. LTF also checks
the current directory and parent directories for various Terraform files
and alters the command line arguments and environment variables to make
Terraform use them.

LTF also executes hooks defined in the first 'ltf.yaml' file it finds
in the current directory or parent directories. This can be used to run
commands or modify the environment before and after Terraform runs.`

func ltf(cwd string, args *arguments, env []string) (cmd *exec.Cmd, exitStatus int) {
	// Special mode to output environment variables after running a hook script.
	// It outputs in JSON format to avoid issues with multi-line variables.
	if len(args.cli) > 1 && args.cli[1] == "-env-to-json" {
		if envJsonBytes, err := json.Marshal(env); err != nil {
			fmt.Fprintf(os.Stderr, "%s: error in env-to-json: %s\n", args.bin, err)
			return nil, 1
		} else {
			fmt.Print(string(envJsonBytes))
			return nil, 0
		}
	}

	// Find and load the optional settings file.
	settings, err := loadSettings(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error loading ltf settings: %s\n", args.bin, err)
		return nil, 1
	}

	// Build the Terraform command.
	cmd, vars, err := terraformCommand(cwd, args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error building command: %s\n", args.bin, err)
		return nil, 1
	}

	// Display variables.
	for _, v := range vars {
		if v.value != "" {
			v.print()
		}
	}

	// Run any "before" hooks.
	if err := settings.runHooks("before", cmd, args, vars); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error from hook: %s\n", args.bin, err)
		return nil, 1
	}

	// Special cases to print messages before Terraform runs.
	if args.help {
		fmt.Println(helpMessage)
		fmt.Println("")
	} else if args.version {
		fmt.Printf("LTF %s\n\n", getVersion())
	}

	// Run the Terraform command.
	exitCode := 0
	if v := getEnvValue(env, "LTF_TEST_MODE"); v != "" {
		fmt.Fprintf(os.Stderr, "# LTF_TEST_MODE=%s skipped %s\n", v, strings.Join(cmd.Args, " "))
	} else {
		fmt.Fprintf(os.Stderr, "# %s\n", strings.Join(cmd.Args, " "))
		if err := cmd.Run(); err != nil {
			if exitErr, isExitError := err.(*exec.ExitError); isExitError {
				exitCode = exitErr.ExitCode()
			} else {
				fmt.Fprintf(os.Stderr, "%s: error running command: %s\n", args.bin, err)
				exitCode = 1
			}
		}
	}

	// Run any "after" or "failed" hooks.
	when := "after"
	if exitCode != 0 {
		when = "failed"
	}
	if err = settings.runHooks(when, cmd, args, vars); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error from hook: %s\n", args.bin, err)
		return nil, 1
	}

	return cmd, exitCode
}

func setDataDir(cmd *exec.Cmd, cwd string, chdir string) error {
	rel, err := filepath.Rel(chdir, cwd)
	if err != nil {
		return err
	}
	env := "TF_DATA_DIR=" + path.Join(rel, ".terraform")
	cmd.Env = append(cmd.Env, env)
	fmt.Fprintf(os.Stderr, "+ %s\n", env)
	return nil
}

func terraformCommand(cwd string, args *arguments, env []string) (cmd *exec.Cmd, vars map[string]*variable, err error) {
	// Start building the Terraform command to run.
	cmd = exec.Command("terraform")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Skip LTF for certain commands and just run Terraform.
	if args.help || args.version || args.subcommand == "" || args.subcommand == "fmt" || args.subcommand == "version" {
		cmd.Args = append(cmd.Args, args.cli[1:]...)
		return cmd, nil, nil
	}

	// Determine the directories to use.
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return nil, nil, err
	}
	dirs, chdir, err := findDirs(cwd, args)
	if err != nil {
		return nil, nil, err
	}

	// Make Terraform change to the configuration directory
	// using the -chdir argument.
	if chdir != cwd && args.chdir == "" {
		if rel, err := filepath.Rel(cwd, chdir); err != nil {
			return nil, nil, err
		} else {
			cmd.Args = append(cmd.Args, "-chdir="+rel)
		}
	}

	// Set the data directory to the current directory.
	if chdir != cwd && getEnvValue(env, "TF_DATA_DIR") == "" {
		if err := setDataDir(cmd, cwd, chdir); err != nil {
			return nil, nil, err
		}
	}

	// Load variables from all possible sources.
	vars, err = loadVariables(args, dirs, chdir)
	if err != nil {
		return nil, nil, err
	}

	// Export loaded variables as environment variables.
	for name, v := range vars {
		env := "TF_VAR_" + name + "=" + v.value
		cmd.Env = append(cmd.Env, env)
	}

	// Use backend files.
	if args.subcommand == "init" {
		if backendConfig, err := loadBackendConfiguration(dirs, chdir, vars); err != nil {
			return nil, nil, err
		} else if len(backendConfig) > 0 {
			// Build the -backend-config arguments.
			args := []string{}
			for name, value := range backendConfig {
				args = append(args, "-backend-config="+name+"="+value)
			}

			// Append the old TF_CLI_ARGS_init at the end so they take precedence
			// over the values generated by LTF.
			old := getEnvValue(env, "TF_CLI_ARGS_init")
			if old != "" {
				args = append(args, old)
			}

			// Set the new environment variable value.
			newEnvValue := strings.Join(args, " ")
			cmd.Env = setEnvValue(cmd.Env, "TF_CLI_ARGS_init", newEnvValue)
			fmt.Fprintf(os.Stderr, "+ TF_CLI_ARGS_init=%s\n", newEnvValue)
		}
	}

	// Pass all command line arguments to Terraform.
	cmd.Args = append(cmd.Args, args.cli[1:]...)

	return cmd, vars, nil
}
