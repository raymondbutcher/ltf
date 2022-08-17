package terraform

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/shlex"
	"github.com/raymondbutcher/ltf"
	"github.com/raymondbutcher/ltf/internal/environ"
)

// NewArguments populates and returns an arguments struct.
func NewArguments(args []string, env ltf.Environ) (*ltf.Arguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments")
	}

	a := ltf.Arguments{}
	a.Args = args
	a.Bin = args[0]
	a.EnvToJson = len(args) > 1 && args[1] == "-env-to-json"

	virtual, err := combineArguments(args, env)
	if err != nil {
		return &a, err
	}
	a.Virtual = virtual

	for _, arg := range virtual[1:] {
		if a.Subcommand == "" && len(arg) > 0 && arg[0:1] != "-" {
			a.Subcommand = arg
		} else if arg == "-help" {
			a.Help = true
		} else if arg == "-version" {
			a.Version = true
		} else if strings.HasPrefix(arg, "-chdir=") {
			a.Chdir = arg[7:]
		}
	}

	// Version can be called as a flag (handled above) or as a subcommand.
	if a.Subcommand == "version" {
		a.Version = true
	}

	return &a, err
}

// cleanArguments converts `-var value` and `-var-file value` arguments
// into `-var=value` and `-var-file=value` respectively.
func cleanArguments(args []string) []string {
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

// combineArguments returns the combined arguments from the CLI arguments
// and the TF_CLI_ARGS and TF_CLI_ARGS_name environment variables.
func combineArguments(args []string, env *ltf.Environ) ([]string, error) {
	args = cleanArguments(args)

	result := []string{args[0]}
	subcommand := ""
	afterEnvArgs := []string{}

	for _, arg := range args[1:] {
		if subcommand == "" {
			result = append(result, arg)
			if arg[0:1] != "-" {
				subcommand = arg
			}
		} else {
			afterEnvArgs = append(afterEnvArgs, arg)
		}
	}

	if envArgs, err := shlex.Split(environ.GetValue(env, "TF_CLI_ARGS")); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", "TF_CLI_ARGS", err)
	} else {
		envArgs = cleanArguments(envArgs)
		for _, arg := range envArgs {
			if subcommand == "" && arg[0:1] != "-" {
				subcommand = arg
			}
			result = append(result, arg)
		}
	}

	if subcommand != "" {
		envName := "TF_CLI_ARGS_" + subcommand
		if envArgs, err := shlex.Split(environ.GetValue(env, envName)); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", envName, err)
		} else {
			envArgs = cleanArguments(envArgs)
			result = append(result, envArgs...)
		}
	}

	result = append(result, afterEnvArgs...)

	return result, nil
}
