package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/shlex"
)

// arguments contain the raw CLI arguments, plus the "virtual" arguments
// which take into account the TF_CLI_ARGS and TF_CLI_ARGS_name environment variables,
// plus some extra useful information.
type arguments struct {
	bin        string
	chdir      string
	cli        []string
	help       bool
	subcommand string
	version    bool
	virtual    []string
}

// newArguments populates and returns an arguments struct.
func newArguments(args []string, env []string) (*arguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments")
	}

	a := arguments{}
	a.bin = args[0]
	a.cli = args

	virtual, err := getVirtualArgs(args, env)
	if err != nil {
		return &a, err
	}
	a.virtual = virtual

	for _, arg := range virtual[1:] {
		if a.subcommand == "" && len(arg) > 0 && arg[0:1] != "-" {
			a.subcommand = arg
		} else if arg == "-help" {
			a.help = true
		} else if arg == "-version" {
			a.version = true
		} else if strings.HasPrefix(arg, "-chdir=") {
			a.chdir = arg[7:]
		}
	}

	// Version can be called as a flag (handled above) or a command.
	if a.subcommand == "version" {
		a.version = true
	}

	return &a, err
}

// cleanArgs converts `-var value` and `-var-file value` arguments
// into `-var=value` and `-var-file=value` respectively.
func cleanArgs(args []string) []string {
	result := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == "-var" || arg == "-var-file") && i < len(args)-1 {
			result = append(result, arg+"="+args[i+1])
			i = i + 1
		} else {
			result = append(result, arg)
		}
	}
	return result
}

// getVirtualArgs returns the combined arguments from the CLI arguments
// and the TF_CLI_ARGS and TF_CLI_ARGS_name environment variables.
func getVirtualArgs(args []string, env []string) ([]string, error) {
	args = cleanArgs(args)

	result := []string{args[0]}

	subcommand := ""
	afterSubcommand := []string{}
	for _, arg := range args[1:] {
		if subcommand == "" {
			result = append(result, arg)
			if arg[0:1] != "-" {
				subcommand = arg
			}
		} else {
			afterSubcommand = append(afterSubcommand, arg)
		}
	}

	envArgs, err := shlex.Split(getEnvValue(env, "TF_CLI_ARGS"))
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", "TF_CLI_ARGS", err)
	}
	for _, arg := range envArgs {
		if subcommand == "" && arg[0:1] != "-" {
			subcommand = arg
		}
		afterSubcommand = append(afterSubcommand, arg)
	}

	result = append(result, afterSubcommand...)

	if subcommand != "" {
		envName := "TF_CLI_ARGS_" + subcommand
		envArgs, err := shlex.Split(getEnvValue(env, envName))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", envName, err)
		}
		result = append(result, envArgs...)
	}

	return result, nil
}
