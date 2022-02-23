package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/tmccombs/hcl2json/convert"
)

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
