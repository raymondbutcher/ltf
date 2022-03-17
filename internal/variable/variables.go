package variable

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/raymondbutcher/ltf/internal/arguments"
	"github.com/raymondbutcher/ltf/internal/environ"
	"github.com/raymondbutcher/ltf/internal/filesystem"
	"github.com/tmccombs/hcl2json/convert"
)

type Variables map[string]*Variable

// SetValue adds or updates a variable. If freeze is true, it sets the variable to Frozen,
// and is able to update existing frozen variables. If freeze is false, and there is an
// existing frozen variable with a different value, it will error.
func (vars Variables) SetValue(name string, value string, freeze bool) (v *Variable, err error) {
	var found bool

	v, found = vars[name]
	if !found {
		if v, err = New(name, "", value); err != nil {
			return nil, err
		}
		vars[name] = v
	}

	if !freeze && v.Frozen && v.StringValue != value {
		return nil, fmt.Errorf("cannot change frozen variable %s", name)
	}

	if err := v.SetValue(value); err != nil {
		return nil, err
	}

	if freeze {
		v.Frozen = true
	}

	return v, nil
}

// SetValues sets multiple variable values. It uses the same freeze logic as SetValue.
func (vars Variables) SetValues(values map[string]string, freeze bool) error {
	for name, value := range values {
		if _, err := vars.SetValue(name, value, freeze); err != nil {
			return err
		}
	}
	return nil
}

// Load returns variables from CLI arguments, environment variables,
// the Terraform configuration, and tfvars files.
//
// Note about value types, which LTF adheres to:
//   https://www.terraform.io/language/values/variables#complex-typed-values
// For convenience, Terraform defaults to interpreting -var and environment
// variable values as literal strings... However, if a root module variable
// uses a type constraint to require a complex value (list, set, map, object,
// or tuple), Terraform will instead attempt to parse its value using the same
// syntax used within variable definitions files...
func Load(args *arguments.Arguments, dirs []string, chdir string) (vars Variables, err error) {
	vars = Variables{}

	// Parse the Terraform config to get variable types and defaults.
	module, diags := tfconfig.LoadModule(chdir)
	if err := diags.Err(); err != nil {
		return nil, err
	}
	for _, v := range module.Variables {
		value := ""
		if v.Default != nil {
			value, err = environ.MarshalValue(v.Default)
			if err != nil {
				return nil, fmt.Errorf("loading %s default value: %w", v.Name, err)
			}
		}
		nv, err := New(v.Name, v.Type, value)
		if err != nil {
			return nil, fmt.Errorf("loading %s variable: %w", v.Name, err)
		}
		nv.Sensitive = v.Sensitive
		vars[v.Name] = nv
	}

	// Load variables that environment variables will not able to override
	// due to Terraform's variables precedence rules.
	// These will be considered "frozen" values.

	// Load tfvars from the configuration directory.
	// Terraform will use these values over TF_VAR_name so freeze them.
	if v, err := readVariablesDir(chdir); err != nil {
		return nil, err
	} else {
		if err := vars.SetValues(v, true); err != nil {
			return nil, err
		}
	}

	// Load variables from CLI arguments.
	// Terraform will prefer these values over TF_VAR_name so freeze them
	// so LTF can return an error if something tries to set a different
	// value using TF_VAR_name.
	if v, err := readVariablesArgs(args.Virtual); err != nil {
		return nil, err
	} else {
		for name, value := range v {
			if _, err := vars.SetValue(name, value, true); err != nil {
				return nil, err
			}
		}
	}

	// Load variables from *.tfvars and *.tfvars.json files.
	// Use directories in reverse order so variables in deeper directories
	// overwrite variables in parent directories.
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		if dir == chdir {
			// Files in chdir were handled earlier.
			continue
		}
		if v, err := readVariablesDir(dir); err != nil {
			return nil, err
		} else {
			for name, value := range v {
				if _, err := vars.SetValue(name, value, false); err != nil {
					return nil, fmt.Errorf("loading from dir %s: %w", dir, err)
				}
			}
		}
	}

	return vars, nil
}

func filterVariableFiles(files []string) (matches []string) {
	// Returns variables files in the correct order of precedence.
	// https://www.terraform.io/language/values/variables#variable-definition-precedence

	sort.Strings(files)

	// 1. The terraform.tfvars file, if present.
	// 2. The terraform.tfvars.json file, if present.
	autoFiles := []string{}
	for _, name := range files {
		if name == "terraform.tfvars" || name == "terraform.tfvars.json" {
			matches = append(matches, name)
		} else if matched, _ := path.Match("*.auto.tfvars", name); matched {
			autoFiles = append(autoFiles, name)
		} else if matched, _ := path.Match("*.auto.tfvars.json", name); matched {
			autoFiles = append(autoFiles, name)
		}
	}

	// 3. Any *.auto.tfvars or *.auto.tfvars.json files,
	//    processed in lexical order of their filenames.
	matches = append(matches, autoFiles...)

	return matches
}

func readVariablesArgs(args []string) (map[string]string, error) {
	result := map[string]string{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-var=") {
			s := strings.SplitN(arg, "=", 3)
			if len(s) != 3 {
				return nil, fmt.Errorf("invalid argument: %s", arg)
			}
			name := s[1]
			value := s[2]
			result[name] = value
		} else if strings.HasPrefix(arg, "-var-file=") {
			s := strings.SplitN(arg, "=", 2)
			if len(s) != 2 {
				return nil, fmt.Errorf("invalid argument: %s", arg)
			}
			file := s[1]
			v, err := readVariablesFile(file)
			if err != nil {
				return nil, err
			}
			for name, value := range v {
				result[name] = value
			}
		}
	}
	return result, nil
}

func readVariablesDir(dir string) (map[string]string, error) {
	result := map[string]string{}

	files, err := filesystem.ReadNames(dir)
	if err != nil {
		return nil, err
	}

	for _, filename := range filterVariableFiles(files) {
		vars, err := readVariablesFile(path.Join(dir, filename))
		if err != nil {
			return nil, err
		}
		for name, value := range vars {
			result[name] = value
		}
	}

	return result, nil
}

func readVariablesFile(filename string) (map[string]string, error) {
	result := map[string]string{}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var jsonBytes []byte

	if strings.HasSuffix(filename, ".json") {
		jsonBytes = bytes
	} else {
		jsonBytes, err = convert.Bytes(bytes, filename, convert.Options{})
		if err != nil {
			return nil, fmt.Errorf("readVariablesFile converting hcl to json: %w", err)
		}
	}

	vars := map[string]interface{}{}
	if err := json.Unmarshal(jsonBytes, &vars); err != nil {
		return nil, fmt.Errorf("readVariablesFile writing json: %w", err)
	}

	for name, val := range vars {
		env, err := environ.MarshalValue(val)
		if err != nil {
			return nil, fmt.Errorf("readVariablesFile reading json: %w", err)
		}
		result[name] = env
	}

	return result, nil
}
