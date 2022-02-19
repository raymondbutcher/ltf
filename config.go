package main

import (
	"fmt"
	"io/ioutil"
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

func (c *Config) Trigger(when string, cmd *exec.Cmd) {
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
		}
	}
}

func loadConfig(cwd string) (*Config, error) {

	content, err := ioutil.ReadFile("example/ltf.yaml")
	if err != nil {
		return nil, err
	}

	config := Config{}
	if err := yaml.UnmarshalStrict([]byte(content), &config); err != nil {
		return nil, err
	}

	return &config, nil
}
