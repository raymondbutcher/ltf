package environ

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GetValue returns the named environment variable value
// from a list of environment variables in the os.Environ() format.
func GetValue(env []string, name string) string {
	prefix := name + "="
	value := ""
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			value = item[len(prefix):]
		}
	}
	return value
}

// MarshalValue returns returns the JSON encoding of v,
// unless it is a string in which case it returns it as-is.
// The result is suitable for use as a TF_VAR_name environment variable.
func MarshalValue(v interface{}) (string, error) {
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

// SetValue adds an environment variable to a list of environment variables
// in the os.Environ() format. If the variable already exists, it is replaced.
// Returns an updated list of environment variables.
func SetValue(env []string, name string, value string) (modifiedEnv []string) {
	prefix := name + "="
	for _, v := range env {
		if !strings.HasPrefix(v, prefix) {
			modifiedEnv = append(modifiedEnv, v)
		}
	}
	modifiedEnv = append(modifiedEnv, name+"="+value)
	return modifiedEnv
}