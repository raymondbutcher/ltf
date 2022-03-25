package environ

import (
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
