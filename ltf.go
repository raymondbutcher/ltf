package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func findConfigDir(dir string) (string, error) {
	var err error
	var lastDir string
	for {
		// Get the real path of this directory.
		dir, err = filepath.Abs(dir)
		if err != nil {
			return "", fmt.Errorf("error getting absolute path %s: %s", dir, err)
		}

		// Stop if this directory was already checked.
		// This occurs after reaching the filesystem root.
		if dir == lastDir {
			return "", errors.New("checked all parent directories")
		}

		// Get the file names in this directory.
		files, err := getFileNames(dir)
		if err != nil {
			return "", fmt.Errorf("error reading directory %s: %s", dir, err)
		}

		// Return this directory if it contains configuration files.
		if hasConfFile(files) {
			return dir, nil
		}

		// Not found, look in the parent directory next time.
		lastDir = dir
		dir = path.Join(dir, "..")
	}
}

func filterTfbackend(files []string) []string {
	matches := []string{}
	for _, name := range files {
		if matched, _ := path.Match("*.tfbackend", name); matched {
			matches = append(matches, name)
		}
	}
	return matches
}

func filterTfvars(files []string) []string {
	// https://www.terraform.io/language/values/variables#variable-definition-precedence

	matches := []string{}

	// The terraform.tfvars file, if present.
	for _, name := range files {
		if name == "terraform.tfvars" {
			matches = append(matches, name)
		}
	}

	// The terraform.tfvars.json file, if present.
	for _, name := range files {
		if name == "terraform.tfvars.json" {
			matches = append(matches, name)
		}
	}

	// Any *.auto.tfvars or *.auto.tfvars.json files, processed in lexical order of their filenames.
	files = files[:]
	sort.Strings(files)
	for _, name := range files {
		if matched, _ := path.Match("*.auto.tfvars", name); matched {
			matches = append(matches, name)
		}
		if matched, _ := path.Match("*.auto.tfvars.json", name); matched {
			matches = append(matches, name)
		}
	}

	return matches
}

func hasConfFile(files []string) bool {
	for _, name := range files {
		if matched, _ := path.Match("*.tf", name); matched {
			return true
		}
		if matched, _ := path.Match("*.tf.json", name); matched {
			return true
		}
	}
	return false
}

func terraformCommand(cwd string, args []string, env []string) *exec.Cmd {
	// Start building the Terraform command to run.
	cmd := exec.Command("terraform")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Determine the Terraform configuration directory to use.
	// If the -chdir argument was provided then use that.
	confDir := getNamedArg(args, "chdir")
	if confDir == "" {
		// Find the closest directory with Terraform configuration files.
		// This can be the current directory or any parent directory.
		var findErr error
		confDir, findErr = findConfigDir(cwd)
		if findErr != nil {
			fmt.Fprintf(os.Stderr, "LTF: error finding Terraform configuration files: %s\n", findErr)
			os.Exit(1)
		}

		// Make Terraform change to the configuration directory
		// by adding the -chdir argument to the Terraform command.
		if confDir != cwd {
			rel, err := filepath.Rel(cwd, confDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "LTF: error resolving chdir: %s\n", err)
				os.Exit(1)
			}
			cmd.Args = append(cmd.Args, fmt.Sprintf("-chdir=%s", rel))
		}
	}

	// Keep the data directory inside the current directory
	// unless the TF_DATA_DIR environment variable is already set.
	if confDir != cwd && getEnv(env, "TF_DATA_DIR") == "" {
		rel, err := filepath.Rel(confDir, cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LTF: error resolving TF_DATA_DIR: %s\n", err)
			os.Exit(1)
		}
		dataDir := path.Join(rel, ".terraform")
		env := fmt.Sprintf("TF_DATA_DIR=%s", dataDir)
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
	}

	// Read the directory.
	cwdFiles, err := getFileNames(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LTF: error reading current working directory %s: %s\n", cwd, err)
		os.Exit(1)
	}

	// Look in current directory for tfbackend files to use.
	// TODO: look in parent directories too, stopping at confiDir.
	backendFiles := filterTfbackend(cwdFiles)
	if len(backendFiles) > 0 {
		argValues := []string{}
		argValue := getEnv(env, "TF_CLI_ARGS_init")
		if argValue != "" {
			argValues = append(argValues, argValue)
		}
		for _, name := range backendFiles {
			abs := path.Join(cwd, name)
			rel, err := filepath.Rel(confDir, abs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "LTF: error resolving relative backend path for %s from %s: %s\n", abs, confDir, err)
				os.Exit(1)
			}
			argValues = append(argValues, fmt.Sprintf("-backend-config=%s", rel))
		}
		env := fmt.Sprintf("TF_CLI_ARGS_init=%s", strings.Join(argValues, " "))
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
	}

	// Look in current directory for tfvars files to automatically use.
	// TODO: look in parent directories too, stopping at confiDir.
	tfvarsFiles := filterTfvars(cwdFiles)
	if len(tfvarsFiles) > 0 {
		for _, argName := range []string{"plan", "apply"} {
			envName := fmt.Sprintf("TF_CLI_ARGS_%s", argName)
			argValues := []string{}
			argValue := getEnv(env, envName)
			if argValue != "" {
				argValues = append(argValues, argValue)
			}
			for _, name := range tfvarsFiles {
				abs := path.Join(cwd, name)
				rel, err := filepath.Rel(confDir, abs)
				if err != nil {
					fmt.Fprintf(os.Stderr, "LTF: error resolving relative tfvars path for %s from %s: %s\n", abs, confDir, err)
					os.Exit(1)
				}
				argValues = append(argValues, fmt.Sprintf("-var-file=%s", rel))
			}
			env := fmt.Sprintf("%s=%s", envName, strings.Join(argValues, " "))
			cmd.Env = append(cmd.Env, env)
			fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
		}
	}

	// Pass all command line arguments to Terraform.
	cmd.Args = append(cmd.Args, args[1:]...)

	return cmd
}
