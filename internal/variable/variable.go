package variable

import (
	"fmt"
	"os"
)

type Variable struct {
	Name      string
	Value     string
	Sensitive bool
	Frozen    bool
}

func New(name string, value string) *Variable {
	return &Variable{Name: name, Value: value}
}

func (v *Variable) Print() {
	if v.Sensitive {
		fmt.Fprintf(os.Stderr, "+ TF_VAR_%s=%s\n", v.Name, "(sensitive value)")
	} else {
		fmt.Fprintf(os.Stderr, "+ TF_VAR_%s=%s\n", v.Name, v.Value)
	}
}
