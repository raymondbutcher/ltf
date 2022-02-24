package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
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

func command(cwd string, args *arguments, env []string) (cmd *exec.Cmd, frozen map[string]string, err error) {
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

	// Parse the Terraform config to get variable defaults.
	vars := map[string]string{}
	module, diags := tfconfig.LoadModule(chdir)
	if err := diags.Err(); err != nil {
		return nil, nil, err
	}
	for _, variable := range module.Variables {
		if variable.Default != nil {
			if value, err := marshalEnvValue(variable.Default); err != nil {
				return nil, nil, fmt.Errorf("reading configuration: %w", err)
			} else {
				vars[variable.Name] = value
			}
		}
	}

	// Load variables that environment variables will not able to override
	// due to Terraform's variables precedence rules.
	// These will be considered "frozen" values.

	// Load tfvars from the configuration directory.
	// Terraform will use these values over TF_VAR_name so freeze them.
	frozen = map[string]string{}
	if v, err := readVariablesDir(chdir); err != nil {
		return nil, nil, err
	} else {
		for name, value := range v {
			vars[name] = value
			frozen[name] = value
		}
	}

	// Load variables from CLI arguments.
	// Terraform will prefer these values over TF_VAR_name so freeze them
	// so LTF can return an error if something tries to set a different
	// value using TF_VAR_name.
	if v, err := readVariablesArgs(args.virtual); err != nil {
		return nil, nil, err
	} else {
		for name, value := range v {
			vars[name] = value
			frozen[name] = value
		}
	}

	// Load variables from *.tfvars and *.tfvars.json files.
	// Use directories in reverse order so variables in deeper directories
	// overwrite variables in parent directories.
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		if dir == chdir {
			// Files in chdir were handled earlier.
			continue
		}
		if v, err := readVariablesDir(dir); err != nil {
			return nil, nil, err
		} else {
			for name, value := range v {
				if frozenValue, found := frozen[name]; found && value != frozenValue {
					return nil, nil, fmt.Errorf("cannot change frozen variable %s from %s", name, dir)
				}
				vars[name] = value
			}
		}
	}

	// Export loaded variables as environment variables.
	for name, value := range vars {
		env := "TF_VAR_" + name + "=" + value
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "[LTF] %s\n", env)
	}

	// Use backend files.
	if args.subcommand == "init" {
		backendFiles, err := findBackendFiles(dirs, chdir)
		if err != nil {
			return nil, nil, err
		}
		if len(backendFiles) > 0 {
			initArgs := []string{}
			for _, file := range backendFiles {
				if backendConfig, err := parseBackendFile(file, vars); err != nil {
					return nil, nil, err
				} else {
					for name, value := range backendConfig {
						initArgs = append(initArgs, "-backend-config="+name+"="+value)
					}
				}
			}
			original := getEnvValue(env, "TF_CLI_ARGS_init")
			if original != "" {
				initArgs = append(initArgs, original)
			}
			env := "TF_CLI_ARGS_init=" + strings.Join(initArgs, " ")
			cmd.Env = append(cmd.Env, env)
			fmt.Fprintf(os.Stderr, "[LTF] %s\n", env)
		}
	}

	// Pass all command line arguments to Terraform.
	cmd.Args = append(cmd.Args, args.cli[1:]...)

	return cmd, frozen, nil
}

func ltf(cwd string, args *arguments, env []string) (cmd *exec.Cmd, exitStatus int) {
	// Special mode to output environment variables after running a hook script.
	// It outputs in JSON format to avoid issues with multi-line variables.
	if len(args.cli) > 1 && args.cli[1] == "-ltf-env-to-json" {
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
	cmd, frozen, err := command(cwd, args, env)
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
	if args.help {
		fmt.Println(helpMessage)
		fmt.Println("")
	}

	// Run the Terraform command.
	fmt.Fprintf(os.Stderr, "[LTF] Running: %s\n", strings.Join(cmd.Args, " "))
	exitCode := 0
	if getEnvValue(env, "LTF_TEST_MODE") != "" {
		fmt.Fprintf(os.Stderr, "[LTF] Skipped because LTF_TEST_MODE is set\n")
	} else {
		if err := cmd.Run(); err != nil {
			if exitErr, isExitError := err.(*exec.ExitError); isExitError {
				exitCode = exitErr.ExitCode()
			} else {
				fmt.Fprintf(os.Stderr, "[LTF] Error starting Terraform: %s\n", err)
				exitCode = 1
			}
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

func setDataDir(cmd *exec.Cmd, cwd string, chdir string) error {
	rel, err := filepath.Rel(chdir, cwd)
	if err != nil {
		return err
	}
	env := "TF_DATA_DIR=" + path.Join(rel, ".terraform")
	cmd.Env = append(cmd.Env, env)
	fmt.Fprintf(os.Stderr, "[LTF] %s\n", env)
	return nil
}
