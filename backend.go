package main

import (
	"path"
	"sort"
)

// findBackendFiles returns backend files to use in the Terraform command.
func findBackendFiles(dirs []string, chdir string) (backendFiles []string, err error) {

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
