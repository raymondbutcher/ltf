package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

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

func wrapperCommand(cwd string, args []string, env []string) (cmd *exec.Cmd, frozen map[string]string, err error) {
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

	// Start building the Terraform command to run.
	cmd = exec.Command("terraform")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
	if chdir != cwd && getNamedArg(args, "chdir") == "" {
		rel, err := filepath.Rel(cwd, chdir)
		if err != nil {
			return nil, nil, err
		}
		cmd.Args = append(cmd.Args, "-chdir="+rel)
	}

	// Set the data directory to the current directory.
	if chdir != cwd && getEnvValue(env, "TF_DATA_DIR") == "" {
		err := setDataDir(cmd, cwd, chdir)
		if err != nil {
			return nil, nil, err
		}
	}

	// Parse the Terraform config to get variable defaults.
	vars := map[string]string{}
	module, _ := tfconfig.LoadModule(chdir)
	for _, variable := range module.Variables {
		if variable.Default != nil {
			env, err := marshalEnvValue(variable.Default)
			if err != nil {
				return nil, nil, fmt.Errorf("reading configuration: %w", err)

			}
			vars[variable.Name] = env
		}
	}

	// Parse *.tfvars and *.tfvars.json files
	// and export TF_VAR_name environment variables.
	// TODO: also handle TF_CLI_ARGS and -var and -var-file etc
	frozen = map[string]string{}
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		v, err := readVariablesDir(dir)
		if err != nil {
			return nil, nil, err
		}
		for name, value := range v {
			if dir == chdir {
				// Variables in tfvars files in the Terraform configuration directory
				// cannot be overridden by environment variables, so keep track of them.
				frozen[name] = value
			} else {
				// If a tfvars file outside of the configuration directory tries to change
				// a frozen variable, then return an error.
				if frozenValue, found := frozen[name]; found && value != frozenValue {
					return nil, nil, fmt.Errorf("TF_VAR_%s would be ignored because it is defined in a tfvars file in the configuration directory", name)
				}
			}
			vars[name] = value
		}
	}

	// Export parsed variables as environment variables.
	for name, value := range vars {
		env := "TF_VAR_" + name + "=" + value
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "[LTF] %s\n", env)
	}

	// Use backend files.
	backendFiles, err := findBackendFiles(dirs, chdir)
	if err != nil {
		return nil, nil, err
	}
	if len(backendFiles) > 0 {
		initArgs := []string{}
		for _, file := range backendFiles {
			rel, err := filepath.Rel(dirs[len(dirs)-1], file)
			if err != nil {
				return nil, nil, err
			}
			initArgs = append(initArgs, "-backend-config="+rel)
		}
		original := getEnvValue(env, "TF_CLI_ARGS_init")
		if original != "" {
			initArgs = append(initArgs, original)
		}
		env := "TF_CLI_ARGS_init=" + strings.Join(initArgs, " ")
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "[LTF] %s\n", env)
	}

	// Pass all command line arguments to Terraform.
	cmd.Args = append(cmd.Args, args[1:]...)

	return cmd, frozen, nil
}
