package hook

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/raymondbutcher/ltf"
)

type Hooks map[string]*Hook

func (m Hooks) Run(when string, cmd *exec.Cmd, args *ltf.Arguments, vars ltf.VariableService) error {
	for _, h := range m {
		if h.Match(when, args) {
			modifiedEnv, err := h.Run(ltf.NewEnviron(cmd.Env))
			if err != nil {
				return err
			}

			for _, env := range modifiedEnv.ListValues() {
				s := strings.SplitN(env, "=", 2)
				if len(s) == 2 {
					name := s[0]
					if len(name) > 7 && name[:7] == "TF_VAR_" {
						name = name[7:]
						value := s[1]
						v, err := vars.SetValue(name, value, false)
						if err != nil {
							return fmt.Errorf("hook %s: %w", h.Name, err)
						}
						v.Print()
					}
				}
			}
			cmd.Env = modifiedEnv.ListValues()
		}
	}
	return nil
}
