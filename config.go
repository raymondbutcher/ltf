package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

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

const bashScript = `#!/bin/bash
exec 3>&1
exec 1>&2

%s

ltf -ltf-env-to-json >&3
`

func (c *Config) Trigger(when string, cmd *exec.Cmd) error {
	subcommand, _, _ := parseArgs(cmd.Args)
	for name, hook := range c.Hooks {
		hookCmds := []string{}
		if when == "before" {
			hookCmds = hook.Before
		} else if when == "after" {
			hookCmds = hook.Before
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
			fmt.Printf("[LTF] running hook: %s (TODO)\n", name)
			for _, script := range hook.Run {
				// TODO: explain how this works.
				// It is updating the environmet on the command.
				hookCmd := exec.Command("bash", "-c", fmt.Sprintf(bashScript, script))
				hookCmd.Env = cmd.Env
				hookCmd.Stdin = os.Stdin
				hookCmd.Stderr = os.Stderr
				envJsonBytes, err := hookCmd.Output()
				if err != nil {
					return err
				}
				newEnv := []string{}
				err = json.Unmarshal(envJsonBytes, &newEnv)
				if err != nil {
					return err
				}
				cmd.Env = newEnv
				// TODO: it's not passing the env through the commands properly
			}
		}
	}
	return nil
}

func loadConfig(cwd string) (*Config, error) {

	content, err := ioutil.ReadFile("../ltf.yaml")
	if err != nil {
		return nil, err
	}

	config := Config{}
	if err := yaml.UnmarshalStrict([]byte(content), &config); err != nil {
		return nil, err
	}

	return &config, nil
}
