package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// getEnvValue returns the requested environment variable value
// from a list of environment variables returned by os.Environ().
func getEnvValue(env []string, key string) string {
	prefix := key + "="
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
