package filesystem

import (
	"os"
	"path"
	"path/filepath"

	"github.com/raymondbutcher/ltf/internal/arguments"
)

func FindDirs(cwd string, args *arguments.Arguments) (dirs []string, chdir string, err error) {
	// Returns directories to use, including the directory to change to.
	// Subtle: chdir is sometimes cwd and won't be used
	// Subtle: dirs always includes chdir (which may be cwd)

	if args.Chdir != "" {
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
		if files, err = ReadNames(dir); err != nil {
			return nil, err
		} else if len(MatchNames(files, "*.tf")) > 0 || len(MatchNames(files, "*.tf.json")) > 0 {
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

func MatchNames(files []string, pattern string) []string {
	matches := []string{}
	for _, name := range files {
		if matched, _ := path.Match(pattern, name); matched {
			matches = append(matches, name)
		}
	}
	return matches
}

func ReadNames(dir string) ([]string, error) {
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
