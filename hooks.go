package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var scriptPreamble = fmt.Sprintf(`#!/bin/bash
exec 3>&1 # redirect 3 to stdout
exec 1>&2 # redirect stdout to sterr
__ltf_env_to_json () {
  local code=$?
  if [ $code -eq 0 ]; then
	"%s" -ltf-env-to-json >&3
  fi
  trap - EXIT
  exit $code
}
trap __ltf_env_to_json EXIT
`, os.Args[0])

type Hook struct {
	Name   string
	Before []string `yaml:"before"`
	After  []string `yaml:"after"`
	Failed []string `yaml:"failed"`
	Script string   `yaml:"script"`
}

// match reports whether the hook matches the given event and command combination.
func (h *Hook) match(when string, cmd *exec.Cmd) (bool, error) {
	subcommand, _, _, err := parseArgs(cmd.Args, cmd.Env)
	if err != nil {
		return false, err
	}
	hookCmds := []string{}
	if when == "before" {
		hookCmds = h.Before
	} else if when == "after" {
		hookCmds = h.After
	} else if when == "failed" {
		hookCmds = h.Failed
	}
	for _, hookCmd := range hookCmds {
		if hookCmd == "terraform" {
			return true, nil
		} else if hookCmd == "terraform "+subcommand {
			return true, nil
		}
	}
	return false, nil
}

func (h *Hook) run(env []string) (modifiedEnv []string, err error) {
	// Runs the hook script and returns the potentially modified environment variables.

	fmt.Fprintf(os.Stderr, "[LTF] Running hook: %s\n", h.Name)

	hookCmd := exec.Command("bash", "-c", scriptPreamble+h.Script)
	hookCmd.Env = env
	hookCmd.Stdin = os.Stdin
	hookCmd.Stderr = os.Stderr

	stdout, err := hookCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := hookCmd.Start(); err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	if err := hookCmd.Wait(); err != nil {
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
