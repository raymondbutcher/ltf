package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func getEnv(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return item[len(prefix):]
		}
	}
	return ""
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

func getNamedArg(args []string, name string) string {
	prefix := "-" + name + "="
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, prefix) {
			return arg[len(prefix):]
		}
	}
	return ""
}

func getParentDirs(dir string) ([]string, error) {
	// Returns the specified directory and all of its parent directories.

	var err error

	dirs := []string{}

	for {
		// Add this directory.
		dirs = append(dirs, dir)

		// Move to the parent directory.
		dir, err = filepath.Abs(path.Dir(dir))
		if err != nil {
			return nil, err
		}

		// Stop if this directory was already checked.
		// This occurs after reaching the filesystem root.
		if dir == dirs[len(dirs)-1] {
			return dirs, nil
		}
	}
}

func matchFiles(files []string, pattern string) []string {
	matches := []string{}
	for _, name := range files {
		if matched, _ := path.Match(pattern, name); matched {
			matches = append(matches, name)
		}
	}
	return matches
}
