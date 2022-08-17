package ltf

import (
	"strings"
)

type Environ struct {
	env []string
}

func NewEnviron(env []string) Environ {
	return Environ{env}
}

// GetValue returns the named environment variable value.
func (env Environ) GetValue(name string) string {
	prefix := name + "="
	value := ""
	for _, item := range env.env {
		if strings.HasPrefix(item, prefix) {
			value = item[len(prefix):]
		}
	}
	return value
}

func (env *Environ) ListValues() []string {
	return env.env
}

// SetValue adds an environment variable to the environment variables.
// If the variable already exists, it is replaced.
func (env *Environ) SetValue(name string, value string) {
	newEnv := []string{}
	prefix := name + "="
	for _, v := range env.env {
		if !strings.HasPrefix(v, prefix) {
			newEnv = append(newEnv, v)
		}
	}
	newEnv = append(newEnv, name+"="+value)
	env.env = newEnv
}
