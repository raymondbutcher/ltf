package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/tmccombs/hcl2json/convert"
)

func filterVariableFiles(files []string) (matches []string) {
	// Returns variables files in the correct order of precedence.
	// https://www.terraform.io/language/values/variables#variable-definition-precedence

	sort.Strings(files)

	// 1. The terraform.tfvars file, if present.
	// 2. The terraform.tfvars.json file, if present.
	autoFiles := []string{}
	for _, name := range files {
		if name == "terraform.tfvars" || name == "terraform.tfvars.json" {
			matches = append(matches, name)
		} else if matched, _ := path.Match("*.auto.tfvars", name); matched {
			autoFiles = append(autoFiles, name)
		} else if matched, _ := path.Match("*.auto.tfvars.json", name); matched {
			autoFiles = append(autoFiles, name)
		}
	}

	// 3. Any *.auto.tfvars or *.auto.tfvars.json files,
	//    processed in lexical order of their filenames.
	matches = append(matches, autoFiles...)

	return matches
}

func findBackendFiles(dirs []string, chdir string) (backendFiles []string, err error) {
	// Returns backend files to use in the Terraform command.

	// Start at the highest directory (configuration directory)
	// and go deeper towards the current directory.
	// Files in the current directory take precedence
	// over files in parent directories.
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		// Get a sorted list of files in this directory.
		files, err := getFileNames(dir)
		if err != nil {
			return nil, err
		}
		sort.Strings(files)

		// Add any matching backend files.
		for _, name := range matchFiles(files, "*.tfbackend") {
			backendFiles = append(backendFiles, path.Join(dir, name))
		}
	}

	return backendFiles, nil
}

func findDirs(cwd string, args []string) (dirs []string, chdir string, err error) {
	// Returns directories to use, including the directory to change to.
	// Subtle: chdir is sometimes cwd and won't be used
	// Subtle: dirs always includes chdir (which may be cwd)

	chdir = getNamedArg(args, "chdir")
	if chdir != "" {
		// The -chdir argument was provided.
		chdir, err = filepath.Abs(chdir)
		if err != nil {
			return nil, "", err
		}
		// Find directories to use for variables/backend files.
		dirs, err = findDirsWithChdir(cwd, chdir)
		if err != nil {
			return nil, "", err
		}
	} else {
		// Find the configuration directory to use,
		// and directories to use for variables/backend files.
		dirs, err = findDirsWithoutChdir(cwd)
		if err != nil {
			return nil, "", err
		}
		chdir = dirs[len(dirs)-1]
	}
	return dirs, chdir, nil
}

func findDirsWithChdir(cwd string, chdir string) ([]string, error) {
	// Returns directories between the current directory and the specified
	// chdir directory. If the chdir directory is not a parent directory
	// of the current directory, then only the current directory and
	// the chdir directory are returned.

	var err error

	dir := cwd
	dirs := []string{}

	for {
		dirs = append(dirs, dir)

		// Stop if this is chdir directory.
		if dir == chdir {
			return dirs, nil
		}

		// Otherwise, move to the parent directory.
		dir, err = filepath.Abs(path.Dir(dir))
		if err != nil {
			return nil, err
		}

		// Stop if this directory was already checked.
		// This occurs after reaching the filesystem root.
		if dir == dirs[len(dirs)-1] {
			// Because the chdir directory was not found in the parents,
			// return only the current directory and the chdir directory.
			if cwd == chdir {
				return []string{cwd}, nil
			} else {
				return []string{cwd, chdir}, nil
			}
		}
	}
}

func findDirsWithoutChdir(cwd string) ([]string, error) {
	// Returns all directories between the current directory
	// and a parent directory containing Terraform configuration files,
	// which will be used as the configuration directory. If no configuration
	// directory is found, then only the current directory is returned.

	var err error
	var files []string

	dir := cwd
	dirs := []string{}

	for {
		dirs = append(dirs, dir)

		// Stop if this directory contains configuration files.
		if files, err = getFileNames(dir); err != nil {
			return nil, err
		} else if len(matchFiles(files, "*.tf")) > 0 || len(matchFiles(files, "*.tf.json")) > 0 {
			return dirs, nil
		}

		// Otherwise, move to the parent directory.
		dir, err = filepath.Abs(path.Dir(dir))
		if err != nil {
			return nil, err
		}

		// Stop if this directory was already checked.
		// This occurs after reaching the filesystem root.
		if dir == dirs[len(dirs)-1] {
			// Because no configuration directory was found,
			// return only the current directory.
			return []string{cwd}, nil
		}
	}
}

func readVariablesFile(filename string) (map[string]string, error) {
	result := map[string]string{}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var jsonBytes []byte

	if strings.HasSuffix(filename, ".json") {
		jsonBytes = bytes
	} else {
		jsonBytes, err = convert.Bytes(bytes, filename, convert.Options{})
		if err != nil {
			return nil, fmt.Errorf("convert hcl to json: %w", err)
		}
	}

	v := map[string]interface{}{}
	if err := json.Unmarshal(jsonBytes, &v); err != nil {
		return nil, fmt.Errorf("unmarshal json from hcl: %w", err)
	}

	for name, val := range v {
		// TODO: strings have unnecessary quotes
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("marshal variable to json: %w", err)
		}
		result[name] = string(jsonBytes)
	}

	return result, nil
}

func readVariablesDir(dir string) (map[string]string, error) {
	result := map[string]string{}

	files, err := getFileNames(dir)
	if err != nil {
		return nil, err
	}

	for _, filename := range filterVariableFiles(files) {
		vars, err := readVariablesFile(path.Join(dir, filename))
		if err != nil {
			return nil, err
		}
		for name, value := range vars {
			result[name] = value
		}
	}

	return result, nil
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

func wrapperCommand(cwd string, args []string, env []string) (*exec.Cmd, error) {
	var err error

	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return nil, err
	}

	// Start building the Terraform command to run.
	cmd := exec.Command("terraform")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Determine the directories to use.
	dirs, chdir, err := findDirs(cwd, args)
	if err != nil {
		return nil, err
	}

	// Make Terraform change to the configuration directory
	// using the -chdir argument.
	if chdir != cwd && getNamedArg(args, "chdir") == "" {
		rel, err := filepath.Rel(cwd, chdir)
		if err != nil {
			return nil, err
		}
		cmd.Args = append(cmd.Args, "-chdir="+rel)
	}

	// Set the data directory to the current directory.
	if chdir != cwd && getEnv(env, "TF_DATA_DIR") == "" {
		err := setDataDir(cmd, cwd, chdir)
		if err != nil {
			return nil, err
		}
	}

	// Parse the Terraform config to get variable defaults.
	vars := map[string]string{}
	module, _ := tfconfig.LoadModule(chdir)
	for _, variable := range module.Variables {
		if variable.Default != nil {
			// TODO: strings have unnecessary quotes
			jsonBytes, err := json.Marshal(variable.Default)
			if err != nil {
				return nil, fmt.Errorf("marshal variable to json: %w", err)
			}
			vars[variable.Name] = string(jsonBytes)
		}
	}

	// Parse *.tfvars and *.tfvars.json files
	// and export TF_VAR_name environment variables.
	// TODO: also handle TF_CLI_ARGS and -var and -var-file etc
	frozen := map[string]string{}
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		v, err := readVariablesDir(dir)
		if err != nil {
			return nil, err
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
					return nil, fmt.Errorf("variable %s found in configuration directory and other directory", name)
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
		return nil, err
	}
	if len(backendFiles) > 0 {
		initArgs := []string{}
		for _, file := range backendFiles {
			rel, err := filepath.Rel(dirs[len(dirs)-1], file)
			if err != nil {
				return nil, err
			}
			initArgs = append(initArgs, "-backend-config="+rel)
		}
		original := getEnv(env, "TF_CLI_ARGS_init")
		if original != "" {
			initArgs = append(initArgs, original)
		}
		env := "TF_CLI_ARGS_init=" + strings.Join(initArgs, " ")
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "[LTF] %s\n", env)
	}

	// Pass all command line arguments to Terraform.
	cmd.Args = append(cmd.Args, args[1:]...)

	return cmd, nil
}
