package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func findConfigDir(dir string, files []string) (string, error) {
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

		// Return this directory if it contains configuration files.
		if hasConfFile(files) {
			return dir, nil
		}

		// Not found, look in the parent directory next time.
		lastDir = dir
		dir = path.Join(dir, "..")
		files, err = getFileNames(dir)
		if err != nil {
			return "", fmt.Errorf("error reading directory %s: %s", dir, err)
		}
	}
}

func getFileNames(dir string) ([]string, error) {
	file, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	names, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	return names, err
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

func hasVarsFile(files []string) bool {
	for _, name := range files {
		if matched, _ := path.Match("*.tfvars", name); matched {
			return true
		}
		if matched, _ := path.Match("*.tfvars.json", name); matched {
			return true
		}
	}
	return false
}

func main() {
	// Prevent users from using the -chdir command line argument.
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-chdir=") {
			fmt.Fprintf(os.Stderr, "Invalid argument for ltf: -chdir\n\n")
			fmt.Fprintf(os.Stderr, "ltf decides the value for -chdir when it runs terraform,\n")
			fmt.Fprintf(os.Stderr, "so -chdir cannot be used as a command line argument for ltf.\n")
			os.Exit(1)
		}
	}

	// Read the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading working directory: %s\n", err)
		os.Exit(1)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path %s: %s\n", cwd, err)
		os.Exit(1)
	}
	files, err := getFileNames(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing files %s: %s\n", cwd, err)
		os.Exit(1)
	}

	// Prevent users from running in a directory with no Terraform files of any kind.
	// This checks for both configuration and variable files.
	if !hasConfFile(files) && !hasVarsFile(files) {
		fmt.Fprint(os.Stderr, "Could not find Terraform variables or configuration files in the current directory\n")
		os.Exit(1)
	}

	// Locate the directory with Terraform configuration files.
	// This can be the current directory or any parent directory.
	confDir, err := findConfigDir(cwd, files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding Terraform configuration files: %s\n", err)
		os.Exit(1)
	}

	// Start building the Terraform command to run.
	cmd := exec.Command("terraform")
	cmd.Env = os.Environ()

	if confDir != cwd {
		// Make Terraform change to the configuration directory.
		cmd.Args = append(cmd.Args, fmt.Sprintf("-chdir=%s", confDir))

		// But keep the data directory inside the current directory.
		dataDir := path.Join(cwd, ".terraform")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TF_DATA_DIR=%s", dataDir))
	}

	// TODO: look in current directory for tfvars files and export them with TF_VAR_name=$value

	// TODO: find backend configuration, render as HCL, add env vars TF_CLI_ARGS_init=-backend-config=$line

	// Pass all other command line arguments to Terraform.
	cmd.Args = append(cmd.Args, os.Args[1:]...)

	// Run the Terraform command.
	if err := cmd.Run(); err != nil {
		if exitErr, isExitError := err.(*exec.ExitError); isExitError {
			os.Exit(exitErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Error running Terraform: %s\n", err)
			os.Exit(1)
		}
	}
}
