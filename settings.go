package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type settings struct {
	Hooks map[string]*hook `yaml:"hooks"`
}

func (s *settings) runHooks(when string, cmd *exec.Cmd, frozen map[string]string) error {
	for _, h := range s.Hooks {
		if matched, err := h.match(when, cmd); err != nil {
			return err
		} else if matched {
			modifiedEnv, err := h.run(cmd.Env)
			if err != nil {
				return err
			}
			for name, value := range frozen {
				if getEnvValue(modifiedEnv, name) != value {
					return fmt.Errorf("cannot change frozen variable %s from hook %s", name, h.Name)
				}
			}
			cmd.Env = modifiedEnv
		}
	}
	return nil
}

func findSettingsFile(dir string) (string, error) {
	// Returns the path to ltf.yaml in the current or parent directories.

	lastDir := ""
	for {
		// Check this directory.
		names, err := getFileNames(dir)
		if err != nil {
			return "", err
		}
		for _, name := range names {
			if name == "ltf.yaml" {
				return path.Join(dir, name), nil
			}
		}

		// Move to the parent directory.
		dir = path.Dir(dir)

		// Stop if this directory was already checked.
		// This occurs after reaching the filesystem root.
		if dir == lastDir {
			break
		}
		lastDir = dir
	}

	// Not found.
	return "", nil
}

func loadSettings(cwd string) (*settings, error) {
	file, err := findSettingsFile(cwd)
	if err != nil {
		return nil, err
	}

	settings := settings{}

	if file == "" {
		return &settings, nil
	}

	rel, err := filepath.Rel(cwd, file)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "[LTF] Loading settings: %s\n", rel)

	content, err := ioutil.ReadFile(rel)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict([]byte(content), &settings); err != nil {
		return nil, err
	}

	for name, hook := range settings.Hooks {
		hook.Name = name
	}

	return &settings, nil
}
