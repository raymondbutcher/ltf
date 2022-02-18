package main

import (
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
			return "", nil
		}

		// Get the file names in this directory.
		files, err := getFileNames(dir)
		if err != nil {
			return "", fmt.Errorf("error reading directory %s: %s", dir, err)
		}

		// Return this directory if it contains configuration files.
		if len(matchFiles(files, "*.tf")) > 0 || len(matchFiles(files, "*.tf.json")) > 0 {
			return dir, nil
		}

		// Not found, look in the parent directory next time.
		lastDir = dir
		dir = path.Join(dir, "..")
	}
}

func matchTfvarsFiles(files []string) []string {
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
	autoMatches := append(matchFiles(files, "*.auto.tfvars"), matchFiles(files, "*.auto.tfvars.json")...)
	sort.Strings(autoMatches)
	matches = append(matches, autoMatches...)

	return matches
}

func terraformCommand(cwd string, args []string, env []string) *exec.Cmd {
	var err error

	// Start building the Terraform command to run.
	cmd := exec.Command("terraform")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Determine the Terraform configuration directory to use.
	// If the -chdir argument was provided then use that.
	// If no configuration files are found then use the current directory.
	confDir := getNamedArg(args, "chdir")
	if confDir == "" {
		// Find the closest directory with Terraform configuration files.
		// This can be the current directory or any parent directory.
		confDir, err = findConfigDir(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LTF: error finding Terraform configuration files: %s\n", err)
			os.Exit(1)
		}
		if confDir == "" {
			confDir = cwd
		}
	}

	// Make Terraform change to the configuration directory
	// by adding the -chdir argument to the Terraform command.
	if confDir != cwd {
		rel, err := filepath.Rel(cwd, confDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LTF: error resolving chdir: %s\n", err)
			os.Exit(1)
		}
		cmd.Args = append(cmd.Args, "-chdir="+rel)
	}

	// Keep the data directory inside the current directory.
	if confDir != cwd && getEnv(env, "TF_DATA_DIR") == "" {
		rel, err := filepath.Rel(confDir, cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LTF: error resolving TF_DATA_DIR: %s\n", err)
			os.Exit(1)
		}
		dataDir := path.Join(rel, ".terraform")
		env := "TF_DATA_DIR=" + dataDir
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
	backendFiles := matchFiles(cwdFiles, "*.tfbackend")
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
			argValues = append(argValues, "-backend-config="+rel)
		}
		env := "TF_CLI_ARGS_init=" + strings.Join(argValues, " ")
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
	}

	// Look in current directory for tfvars files to automatically use.
	// TODO: look in parent directories too, stopping at confiDir.
	tfvarsFiles := matchTfvarsFiles(cwdFiles)
	if len(tfvarsFiles) > 0 {
		for _, argName := range []string{"plan", "apply"} {
			envName := "TF_CLI_ARGS_" + argName
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
				argValues = append(argValues, "-var-file="+rel)
			}
			env := envName + "=" + strings.Join(argValues, " ")
			cmd.Env = append(cmd.Env, env)
			fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
		}
	}

	// Pass all command line arguments to Terraform.
	cmd.Args = append(cmd.Args, args[1:]...)

	return cmd
}
