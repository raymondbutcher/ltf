package ltf

import (
	"fmt"
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

type Variable struct {
	Name        string
	Type        string
	AnyValue    cty.Value
	StringValue string
	Sensitive   bool
	Frozen      bool
}

func NewVariable(name string, vtype string, value string) (*Variable, error) {
	v := &Variable{}
	v.Name = name
	v.Type = vtype
	err := v.SetValue(value)
	return v, err
}

func (v *Variable) Print() {
	if v.Sensitive {
		fmt.Fprintf(os.Stderr, "+ TF_VAR_%s=%s\n", v.Name, "(sensitive value)")
	} else {
		fmt.Fprintf(os.Stderr, "+ TF_VAR_%s=%s\n", v.Name, v.StringValue)
	}
}
func (v *Variable) SetValue(value string) error {
	if v.Type == "" || v.Type == "string" {
		v.AnyValue = cty.StringVal(value)
	} else if value == "" {
		v.AnyValue = cty.NilVal
	} else {
		j := json.SimpleJSONValue{}
		if err := j.UnmarshalJSON([]byte(value)); err != nil {
			return fmt.Errorf("parsing variable value: %w", err)
		}
		v.AnyValue = j.Value
	}
	v.StringValue = value
	return nil
}

// VariableService represents a service for managing variables.
type VariableService interface {
	Each() []*Variable
	GetValue(name string) string
	GetCtyValue(name string) cty.Value
	Load(args *Arguments, dirs []string, chdir string) error
	SetValue(name string, value string, freeze bool) (v *Variable, err error)
	SetValues(values map[string]string, freeze bool) error
}
