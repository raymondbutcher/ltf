package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/tmccombs/hcl2json/convert"
)

type variables map[string]*variable

type variable struct {
	name      string
	value     string
	sensitive bool
	frozen    bool
}

func newVariable(name string, value string) *variable {
	return &variable{name: name, value: value}
}

func (v *variable) print() {
	if v.sensitive {
		fmt.Fprintf(os.Stderr, "+ TF_VAR_%s=%s\n", v.name, "(sensitive value)")
	} else {
		fmt.Fprintf(os.Stderr, "+ TF_VAR_%s=%s\n", v.name, v.value)
	}
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

func loadVariables(args *arguments, dirs []string, chdir string) (vars variables, err error) {
	vars = variables{}

	// Parse the Terraform config to get variable defaults.
	module, diags := tfconfig.LoadModule(chdir)
	if err := diags.Err(); err != nil {
		return nil, err
	}
	for _, v := range module.Variables {
		nv := newVariable(v.Name, "")
		nv.sensitive = v.Sensitive
		if v.Default != nil {
			nv.value, err = marshalEnvValue(v.Default)
			if err != nil {
				return nil, fmt.Errorf("reading configuration: %w", err)
			}
		}
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
		for name, value := range v {
			if _, found := vars[name]; found {
				vars[name].value = value
				vars[name].frozen = true
			} else {
				nv := newVariable(name, value)
				nv.frozen = true
				vars[name] = nv
			}
		}
	}

	// Load variables from CLI arguments.
	// Terraform will prefer these values over TF_VAR_name so freeze them
	// so LTF can return an error if something tries to set a different
	// value using TF_VAR_name.
	if v, err := readVariablesArgs(args.virtual); err != nil {
		return nil, err
	} else {
		for name, value := range v {
			if _, found := vars[name]; found {
				vars[name].value = value
				vars[name].frozen = true
			} else {
				nv := newVariable(name, value)
				nv.frozen = true
				vars[name] = nv
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
				if v, found := vars[name]; found {
					if v.frozen {
						return nil, fmt.Errorf("cannot change frozen variable %s from %s", name, dir)
					}
					v.value = value
				} else {
					vars[name] = newVariable(name, value)
				}
			}
		}
	}

	return vars, nil
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
		env, err := marshalEnvValue(val)
		if err != nil {
			return nil, fmt.Errorf("readVariablesFile reading json: %w", err)
		}
		result[name] = env
	}

	return result, nil
}

func readVariablesDir(dir string) (map[string]string, error) {
	result := map[string]string{}

	files, err := getFileNames(dir)
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
