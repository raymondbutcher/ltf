package arguments

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/shlex"
	"github.com/raymondbutcher/ltf/internal/environ"
)

// Arguments contain the raw CLI Arguments, plus the "virtual" Arguments
// which take into account the TF_CLI_ARGS and TF_CLI_ARGS_name environment variables,
// plus some extra useful information.
type Arguments struct {
	Bin        string
	Chdir      string
	Cli        []string
	Help       bool
	Subcommand string
	Version    bool
	Virtual    []string
}

// New populates and returns an arguments struct.
func New(args []string, env []string) (*Arguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments")
	}

	a := Arguments{}
	a.Bin = args[0]
	a.Cli = args

	virtual, err := getVirtualArgs(args, env)
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
		envArgs = cleanArgs(envArgs)
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
			envArgs = cleanArgs(envArgs)
			result = append(result, envArgs...)
		}
	}

	result = append(result, afterEnvArgs...)

	return result, nil
}
