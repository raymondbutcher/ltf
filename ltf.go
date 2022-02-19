package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func findDirs(cwd string) ([]string, error) {
	cwd, err := filepath.Abs(cwd)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %s", err)
	}

	dir := cwd
	dirs := []string{}

	for {
		dirs = append(dirs, dir)

		// Get the file names in this directory.
		files, err := getFileNames(dir)
		if err != nil {
			return nil, err
		}

		// Stop at this directory if it contains configuration files.
		if len(matchFiles(files, "*.tf")) > 0 || len(matchFiles(files, "*.tf.json")) > 0 {
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
			// Because no configuration directory was found in the parents,
			// return only the current directory.
			return []string{cwd}, nil
		}
	}
}

// func matchVarsFiles(files []string) []string {
// 	// https://www.terraform.io/language/values/variables#variable-definition-precedence

// 	matches := []string{}

// 	// The terraform.tfvars file, if present.
// 	for _, name := range files {
// 		if name == "terraform.tfvars" {
// 			matches = append(matches, name)
// 		}
// 	}

// 	// The terraform.tfvars.json file, if present.
// 	for _, name := range files {
// 		if name == "terraform.tfvars.json" {
// 			matches = append(matches, name)
// 		}
// 	}

// 	// Any *.auto.tfvars or *.auto.tfvars.json files, processed in lexical order of their filenames.
// 	autoMatches := append(matchFiles(files, "*.auto.tfvars"), matchFiles(files, "*.auto.tfvars.json")...)
// 	sort.Strings(autoMatches)
// 	matches = append(matches, autoMatches...)

// 	return matches
// }

func terraformCommand(cwd string, args []string, env []string) (*exec.Cmd, error) {
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
	// Search the current and parent directories until Terraform configuration
	// files are found. If none are found then use the current directory.
	dirs, err := findDirs(cwd)
	if err != nil {
		return nil, err
	}
	chdir := getNamedArg(args, "chdir")
	if chdir != "" {
		// The -chdir argument was provided.
		chdir, err = filepath.Abs(chdir)
		if err != nil {
			return nil, err
		}
		chdirIndex := -1
		for i, dir := range dirs {
			if dir == chdir {
				chdirIndex = i
				break
			}
		}
		if chdirIndex == -1 {
			if chdir == cwd {
				dirs = []string{cwd}
			} else {
				dirs = []string{cwd, chdir}
			}
		} else {
			dirs = dirs[:chdirIndex]
		}
	} else {
		// Make Terraform change to the configuration directory
		// by adding the -chdir argument to the Terraform command.
		chdir = dirs[len(dirs)-1]
		if chdir != cwd {
			rel, err := filepath.Rel(cwd, chdir)
			if err != nil {
				return nil, err
			}
			cmd.Args = append(cmd.Args, "-chdir="+rel)
		}
	}

	// Keep the data directory inside the current directory.
	if chdir != cwd && getEnv(env, "TF_DATA_DIR") == "" {
		rel, err := filepath.Rel(dirs[len(dirs)-1], cwd)
		if err != nil {
			return nil, err
		}
		env := "TF_DATA_DIR=" + path.Join(rel, ".terraform")
		cmd.Env = append(cmd.Env, env)
		fmt.Fprintf(os.Stderr, "LTF: %s\n", env)
	}

	// Find backend and variables files in each directory.
	backendFiles := []string{}
	varFiles := []string{}
	varFileNames := map[string]bool{}
	for _, dir := range dirs {

		files, err := getFileNames(dir)
		if err != nil {
			return nil, err
		}

		for _, name := range matchFiles(files, "*.tfbackend") {
			backendFiles = append(backendFiles, path.Join(dir, name))
		}

		for _, name := range append(matchFiles(files, "*.tfvars"), matchFiles(files, "*.tfvars.json")...) {
			if _, ok := varFileNames[name]; !ok {
				varFiles = append(varFiles, path.Join(dir, name))
				varFileNames[name] = true
			}
		}
	}

	// Use backend files.
	if len(backendFiles) > 0 {
		argValues := []string{}
		argValue := getEnv(env, "TF_CLI_ARGS_init")
		if argValue != "" {
			argValues = append(argValues, argValue)
		}
		for i := len(backendFiles) - 1; i >= 0; i-- {
			file := backendFiles[i]
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
			for i := len(varFiles) - 1; i >= 0; i-- {
				file := varFiles[i]
				rel, err := filepath.Rel(dirs[len(dirs)-1], file)
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
