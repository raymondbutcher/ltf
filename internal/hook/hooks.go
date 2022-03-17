package hook

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/raymondbutcher/ltf/internal/arguments"
	"github.com/raymondbutcher/ltf/internal/variable"
)

type Hooks map[string]*Hook

func (m Hooks) Run(when string, cmd *exec.Cmd, args *arguments.Arguments, vars variable.Variables) error {
	for _, h := range m {
		if h.Match(when, args) {
			modifiedEnv, err := h.Run(cmd.Env)
			if err != nil {
				return err
			}

			for _, env := range modifiedEnv {
				s := strings.SplitN(env, "=", 2)
				if len(s) == 2 {
					name := s[0]
					if len(name) > 7 && name[:7] == "TF_VAR_" {
						name = name[7:]
						value := s[1]
						if v, found := vars[name]; found {
							if value != v.Value {
								if v.Frozen {
									return fmt.Errorf("cannot change frozen variable %s from hook %s", name, h.Name)
								}
								v.Print()
							}
						} else {
							v = variable.New(name, value)
							vars[name] = v
							v.Print()
						}
					}
				}
			}
			cmd.Env = modifiedEnv
		}
	}
	return nil
}
