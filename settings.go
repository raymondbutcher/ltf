package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type settings struct {
	Hooks map[string]*hook `yaml:"hooks"`
}

func (s *settings) runHooks(when string, cmd *exec.Cmd, args *arguments, vars map[string]*variable) error {
	for _, h := range s.Hooks {
		if h.match(when, args) {
			modifiedEnv, err := h.run(cmd.Env)
			if err != nil {
				return err
			}

			for _, env := range modifiedEnv {
				s := strings.SplitN(env, "=", 2)
				if len(s) == 2 {
					name := s[0]
					if len(name) > 7 && name[:7] == "TF_VAR_" {
						name = name[7:]
						value := s[1]
						if v, found := vars[name]; found {
							if value != v.value {
								if v.frozen {
									return fmt.Errorf("cannot change frozen variable %s from hook %s", name, h.Name)
								}
								v.print()
							}
						} else {
							v = newVariable(name, value)
							vars[name] = v
							v.print()
						}
					}
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
