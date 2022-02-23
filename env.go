package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/shlex"
)

// getArgsWithEnv reads the TF_CLI_ARGS and TF_CLI_ARGS_name
// environment variables and combines them with regular CLI arguments.
func getArgsWithEnv(args []string, env []string) ([]string, error) {
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

// getEnvValue returns the requested environment variable value
// from a list of environment variables returned by os.Environ().
func getEnvValue(env []string, name string) string {
	prefix := name + "="
	value := ""
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			value = item[len(prefix):]
		}
	}
	return value
}

// marshalEnvValue returns returns the JSON encoding of v,
// unless it is a string in which case it returns it as-is.
// The result is suitable for use as an environment variable.
func marshalEnvValue(v interface{}) (string, error) {
	if str, ok := v.(string); ok {
		return str, nil
	} else {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("marshal to environment variable: %w", err)
		}
		return string(jsonBytes), nil
	}
}
