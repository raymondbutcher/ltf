package ltf

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/raymondbutcher/ltf"
	"github.com/raymondbutcher/ltf/internal/arguments"
	"github.com/raymondbutcher/ltf/internal/backend"
	"github.com/raymondbutcher/ltf/internal/filesystem"
	"github.com/raymondbutcher/ltf/internal/hook"
	"github.com/raymondbutcher/ltf/internal/settings"
	"github.com/raymondbutcher/ltf/internal/variable"
)

const helpMessage = `LTF is a transparent wrapper for Terraform; it passes all command line
arguments and environment variables through to Terraform. LTF also checks
the current directory and parent directories for various Terraform files
and alters the command line arguments and environment variables to make
Terraform use them.

LTF also executes hooks defined in the first 'ltf.yaml' file it finds
in the current directory or parent directories. This can be used to run
commands or modify the environment before and after Terraform runs.`

func Run(cwd string, args *arguments.Arguments, env ltf.Environ, version string) (cmd *exec.Cmd, exitStatus int, err error) {
	// Special mode to output environment variables after running a hook script.
	// It outputs in JSON format to avoid issues with multi-line variables.
	if args.EnvToJson {
		envJsonBytes, err := json.Marshal(env)
		if err != nil {
			return nil, 1, fmt.Errorf("%s: error in env-to-json: %w", args.Bin, err)
		}
		fmt.Print(string(envJsonBytes))
		return nil, 0, nil
	}

	// Find and load the optional settings file to get hooks.
	var hooks hook.Hooks
	if s, err := settings.Load(cwd); err != nil {
		return nil, 1, fmt.Errorf("error loading ltf settings: %w", err)
	} else {
		hooks = s.Hooks
	}

	// Skip some chdir and variables functionality for these commands.
	skipMode := args.Help || args.Version || args.Subcommand == "" || args.Subcommand == "fmt"

	// Determine the directories to use.
	dirs := []string{}
	chdir := ""
	if !skipMode {
		cwd, err = filepath.Abs(cwd)
		if err != nil {
			return nil, 1, fmt.Errorf("error reading path: %w", err)
		}
		dirs, chdir, err = filesystem.FindDirs(cwd, args)
		if err != nil {
			return nil, 1, fmt.Errorf("error finding directories: %w", err)
		}
	}

	// Set the data directory to the current directory.
	if !skipMode && env.GetValue("TF_DATA_DIR") == "" && chdir != cwd {
		cwdFromChdir, err := filepath.Rel(chdir, cwd)
		if err != nil {
			return nil, 1, fmt.Errorf("error reading path: %w", err)
		}
		dataDir := path.Join(cwdFromChdir, ".terraform")
		env = env.SetValue("TF_DATA_DIR", dataDir)
		fmt.Fprintf(os.Stderr, "+ TF_DATA_DIR=%s\n", dataDir)
	}

	// Load variables from all possible sources.
	vars := variable.Variables{}
	if !skipMode {
		vars, err = variable.Load(args, dirs, chdir)
		if err != nil {
			return nil, 1, fmt.Errorf("error loading variables: %w", err)
		}
		for _, v := range vars {
			env = env.SetValue("TF_VAR_"+v.Name, v.StringValue)
			if v.StringValue != "" {
				v.Print()
			}
		}
	}

	// Build the Terraform command to run.
	cmd = exec.Command("terraform")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Make Terraform change to the configuration directory
	// using the -chdir argument.
	if !skipMode && args.Chdir == "" && chdir != cwd {
		chdirFromCwd, err := filepath.Rel(cwd, chdir)
		if err != nil {
			return nil, 0, err
		}
		cmd.Args = append(cmd.Args, "-chdir="+chdirFromCwd)
	}

	// Pass all remaining command line arguments to Terraform.
	cmd.Args = append(cmd.Args, args.CommandLineArgs[1:]...)

	// Use backend configuration files.
	if !skipMode && args.Subcommand == "init" {
		backend, err := backend.LoadConfiguration(dirs, chdir, vars)
		if err != nil {
			return nil, 1, err
		}
		if len(backend) > 0 {
			// Build the -backend-config arguments.
			initArgs := []string{}
			for name, value := range backend {
				initArgs = append(initArgs, "-backend-config="+name+"="+value)
			}

			// Append the oldArgs TF_CLI_ARGS_init at the end so they take precedence
			// over the values generated by LTF.
			oldArgs := env.GetValue("TF_CLI_ARGS_init")
			if oldArgs != "" {
				initArgs = append(initArgs, oldArgs)
			}

			// Set the new environment variable value.
			newEnvValue := strings.Join(initArgs, " ")
			env = env.SetValue("TF_CLI_ARGS_init", newEnvValue)
			cmd.Env = env
			fmt.Fprintf(os.Stderr, "+ TF_CLI_ARGS_init=%s\n", newEnvValue)
		}
	}

	// Run any "before" hooks.
	if err := hooks.Run("before", cmd, args, vars); err != nil {
		return nil, 1, fmt.Errorf("error from hook: %w", err)
	}

	// Special cases to print messages before Terraform runs.
	if args.Help {
		fmt.Println(helpMessage)
		fmt.Println("")
	} else if args.Version {
		fmt.Printf("LTF %s\n\n", version)
	}

	// Run the Terraform command.
	exitCode := 0
	cmdString := strings.Join(cmd.Args, " ")
	if v := env.GetValue("LTF_TEST_MODE"); v != "" {
		fmt.Fprintf(os.Stderr, "# LTF_TEST_MODE=%s skipped %s\n", v, cmdString)
	} else {
		fmt.Fprintf(os.Stderr, "# %s\n", cmdString)
		if err := cmd.Run(); err != nil {
			if exitErr, isExitError := err.(*exec.ExitError); isExitError {
				exitCode = exitErr.ExitCode()
			} else {
				return nil, 1, fmt.Errorf("error running command: %w", err)
			}
		}
	}

	// Run any "after" or "failed" hooks.
	when := "after"
	if exitCode != 0 {
		when = "failed"
	}
	if err = hooks.Run(when, cmd, args, vars); err != nil {
		return nil, 1, fmt.Errorf("error from hook: %w", err)
	}

	return cmd, exitCode, nil
}
