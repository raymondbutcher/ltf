package ltf

import (
	"strings"
)

// Environ contains environment variables in the same format returned by os.Environ()
// and used by exec.Command{} but with methods to get and set values more easily.
type Environ []string

// NewEnviron returns a new Environ object
// with the provided environment variables in the format "key=value".
func NewEnviron(env ...string) Environ {
	return append(Environ{}, env...)
}

// GetValue returns the named environment variable value or an empty string if it doesn't exist.
func (env Environ) GetValue(name string) string {
	prefix := name + "="
	value := ""
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			value = item[len(prefix):]
		}
	}
	return value
}

// SetValue adds an environment variable to the environment variables.
// If the variable already exists, it is replaced.
// A new, updated Environ object is returned; it does not modify the existing object.
func (env Environ) SetValue(name string, value string) Environ {
	newEnv := Environ{}
	prefix := name + "="
	for _, v := range env {
		if !strings.HasPrefix(v, prefix) {
			newEnv = append(newEnv, v)
		}
	}
	newEnv = append(newEnv, name+"="+value)
	return newEnv
}
