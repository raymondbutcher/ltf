package arguments

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/shlex"
	"github.com/raymondbutcher/ltf"
)

// Arguments contains information about the arguments passed into the LTF
// program via command line arguments and environment variables.
type Arguments struct {
	// Bin is the command that was run.
	Bin string

	// CommandLineArgs holds command line arguments,
	// including the value of Bin as CommandLineArgs[0].
	CommandLineArgs []string

	// Virtual holds the combined arguments from Args and also extra arguments
	// provided by the `TF_CLI_ARGS` and `TF_CLI_ARGS_name` environment variables.
	Virtual []string

	// EnvToJson is true if the -env-to-json flag was specified.
	// This is a special LTF flag used by the hooks system.
	EnvToJson bool

	// Chdir is the value of the -chdir global option if specified.
	Chdir string

	// Help is true if the -help global option is specified.
	Help bool

	// Version is true if the -version global option is specified
	// or the subcommand is "version".
	Version bool

	// Subcommand is the first non-flag argument if there is one.
	Subcommand string
}

// New populates and returns an arguments struct.
func New(args []string, env ltf.Environ) (*Arguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments")
	}

	a := Arguments{}
	a.CommandLineArgs = args
	a.Bin = args[0]
	a.EnvToJson = len(args) > 1 && args[1] == "-env-to-json"

	virtual, err := combine(args, env)
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

// clean converts `-var value` and `-var-file value` arguments
// into `-var=value` and `-var-file=value` respectively.
func clean(args []string) []string {
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

// combine returns the combined arguments from the CLI arguments
// and the TF_CLI_ARGS and TF_CLI_ARGS_name environment variables.
func combine(args []string, env ltf.Environ) ([]string, error) {
	args = clean(args)

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

	if envArgs, err := shlex.Split(env.GetValue("TF_CLI_ARGS")); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", "TF_CLI_ARGS", err)
	} else {
		envArgs = clean(envArgs)
		for _, arg := range envArgs {
			if subcommand == "" && arg[0:1] != "-" {
				subcommand = arg
			}
			result = append(result, arg)
		}
	}

	if subcommand != "" {
		envName := "TF_CLI_ARGS_" + subcommand
		if envArgs, err := shlex.Split(env.GetValue(envName)); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", envName, err)
		} else {
			envArgs = clean(envArgs)
			result = append(result, envArgs...)
		}
	}

	result = append(result, afterEnvArgs...)

	return result, nil
}
