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

func findFiles(dirs []string, chdir string) (backendFiles []string, varFiles []string, err error) {
	// Returns variables and backend files to use in the Terraform command.

	// Start at the highest directory (configuration directory)
	// and go deeper towards the current directory.
	// Files in the current directory take precedence
	// over files in parent directories.
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		// Get a sorted list of files in this directory.
		files, err := getFileNames(dir)
		if err != nil {
			return nil, nil, err
		}
		sort.Strings(files)

		// Add any matching backend files.
		for _, name := range matchFiles(files, "*.tfbackend") {
			backendFiles = append(backendFiles, path.Join(dir, name))
		}

		// Don't search for variables files in the directory where Terraform
		// will run because Terraform already uses those files by itself.
		if dir != chdir {
			// https://www.terraform.io/language/values/variables#variable-definition-precedence
			// 1. The terraform.tfvars file, if present.
			// 2. The terraform.tfvars.json file, if present.
			autoFiles := []string{}
			for _, name := range files {
				if name == "terraform.tfvars" || name == "terraform.tfvars.json" {
					varFiles = append(varFiles, path.Join(dir, name))
				} else if matched, _ := path.Match("*.auto.tfvars", name); matched {
					autoFiles = append(autoFiles, path.Join(dir, name))
				} else if matched, _ := path.Match("*.auto.tfvars.json", name); matched {
					autoFiles = append(autoFiles, path.Join(dir, name))
				}
			}

			// 3. Any *.auto.tfvars or *.auto.tfvars.json files,
			//    processed in lexical order of their filenames.
			varFiles = append(varFiles, autoFiles...)
		}
	}

	return backendFiles, varFiles, nil
}

func setDataDir(cmd *exec.Cmd, cwd string, chdir string) error {
	rel, err := filepath.Rel(chdir, cwd)
	if err != nil {
		return err
	}
	env := "TF_DATA_DIR=" + path.Join(rel, ".terraform")
	cmd.Env = append(cmd.Env, env)
	fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
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
		setDataDir(cmd, cwd, chdir)
	}

	// Find files to use.
	backendFiles, varFiles, err := findFiles(dirs, chdir)
	if err != nil {
		return nil, err
	}

	// Use backend files.
	if len(backendFiles) > 0 {
		argValues := []string{}
		argValue := getEnv(env, "TF_CLI_ARGS_init")
		if argValue != "" {
			argValues = append(argValues, argValue)
		}
		for _, file := range backendFiles {
			rel, err := filepath.Rel(dirs[len(dirs)-1], file)
			if err != nil {
				return nil, err
			}
			argValues = append(argValues, "-backend-config="+rel)
		}
		env := "TF_CLI_ARGS_init=" + strings.Join(argValues, " ")
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
	}

	// Use variables files.
	if len(varFiles) > 0 {
		for _, argName := range []string{"plan", "apply"} {
			envName := "TF_CLI_ARGS_" + argName
			argValues := []string{}
			argValue := getEnv(env, envName)
			if argValue != "" {
				argValues = append(argValues, argValue)
			}
			for _, file := range varFiles {
				rel, err := filepath.Rel(chdir, file)
				if err != nil {
					return nil, err
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

	return cmd, nil
}
