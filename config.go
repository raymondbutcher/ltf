package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Hooks map[string]Hook `yaml:"hooks"`
}

type Hook struct {
	Before []string `yaml:"before"`
	After  []string `yaml:"after"`
	Failed []string `yaml:"failed"`
	Run    []string `yaml:"run"`
}

const scriptTemplate = `#!/bin/bash
set -euo pipefail

exec 3>&1 # redirect 3 to stdout
exec 1>&2 # redirect stdout to sterr

__ltf_env_to_json () {
  local code=$?
  # Path to LTF.
  "%s" -ltf-env-to-json >&3
  trap - EXIT
  exit "$code"
}
trap __ltf_env_to_json EXIT

# Hook script.
%s
`

func (c *Config) Trigger(when string, cmd *exec.Cmd) error {
	subcommand, _, _ := parseArgs(cmd.Args)
	for name, hook := range c.Hooks {
		hookCmds := []string{}
		if when == "before" {
			hookCmds = hook.Before
		} else if when == "after" {
			hookCmds = hook.After
		} else if when == "failed" {
			hookCmds = hook.Failed
		}
		matched := false
		for _, hookCmd := range hookCmds {
			if hookCmd == "terraform" {
				matched = true
				break
			} else if hookCmd == "terraform "+subcommand {
				matched = true
				break
			}
		}
		if matched {
			fmt.Fprintf(os.Stderr, "[LTF] Running hook: %s\n", name)
			for _, script := range hook.Run {
				newEnv, err := runScript(script, cmd.Env)
				if err != nil {
					return err
				}
				cmd.Env = newEnv
			}
		}
	}
	return nil
}

func findConfig(cwd string) (string, error) {
	// Returns the path to ltf.yaml or ltf.yml in the current or parent directories.

	dirs, err := getParentDirs(cwd)
	if err != nil {
		return "", err
	}

	for _, dir := range dirs {
		names, err := getFileNames(dir)
		if err != nil {
			return "", err
		}

		for _, name := range names {
			if name == "ltf.yaml" || name == "ltf.yml" {
				return path.Join(dir, name), nil
			}
		}
	}

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

		// TODO: include in env for hooks?
		fmt.Fprintf(os.Stderr, "[LTF] Loading configuration: %s\n", rel)

		content, err := ioutil.ReadFile(rel)
		if err != nil {
			return nil, err
		}

		if err := yaml.UnmarshalStrict([]byte(content), &config); err != nil {
			return nil, err
		}
	}

	return &config, nil
}

func runScript(script string, env []string) (modifiedEnv []string, err error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(scriptTemplate, os.Args[0], script))
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	if len(bytes) == 0 {
		return nil, errors.New("wrapper script failed to output environment variables")
	}

	err = json.Unmarshal(bytes, &modifiedEnv)
	if err != nil {
		return nil, err
	}

	return modifiedEnv, nil
}
