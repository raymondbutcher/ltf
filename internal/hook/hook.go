package hook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/raymondbutcher/ltf/internal/arguments"
)

var scriptPreamble = fmt.Sprintf(`#!/bin/bash
exec 3>&1 # redirect 3 to stdout
exec 1>&2 # redirect stdout to sterr
__ltf_env_to_json () {
  local code=$?
  if [ $code -eq 0 ]; then
	"%s" -env-to-json >&3
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

// Match reports whether the hook matches the given event and command combination.
func (h *Hook) Match(when string, args *arguments.Arguments) bool {
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
			return true
		} else if args.Subcommand != "" && hookCmd == "terraform "+args.Subcommand {
			return true
		}
	}
	return false
}

// Run executes the hook script and returns the potentially modified environment variables.
func (h *Hook) Run(env []string) (modifiedEnv []string, err error) {
	fmt.Fprintf(os.Stderr, "# %s\n", h.Name)

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
