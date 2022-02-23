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

// TODO: rename to settings to avoid naming conflict with Terraform Configuration

type Config struct {
	Hooks map[string]*Hook `yaml:"hooks"`
}

func (c *Config) runHooks(when string, cmd *exec.Cmd) error {
	for _, h := range c.Hooks {
		if h.match(when, cmd) {
			modifiedEnv, err := h.run(cmd.Env)
			if err != nil {
				return err
			}
			// TODO: return error if modifying frozen variables
			cmd.Env = modifiedEnv
		}
	}
	return nil
}

func findConfig(dir string) (string, error) {
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

func loadConfig(cwd string) (*Config, error) {

	file, err := findConfig(cwd)
	if err != nil {
		return nil, err
	}

	config := Config{}

	if file != "" {

		rel, err := filepath.Rel(cwd, file)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(os.Stderr, "[LTF] Loading configuration: %s\n", rel)

		content, err := ioutil.ReadFile(rel)
		if err != nil {
			return nil, err
		}

		if err := yaml.UnmarshalStrict([]byte(content), &config); err != nil {
			return nil, err
		}

		for name, hook := range config.Hooks {
			hook.Name = name
		}
	}

	return &config, nil
}
