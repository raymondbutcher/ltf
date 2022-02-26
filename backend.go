package main

import (
	"path"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// findBackendFilenames returns *.tfbackend files to use for the Terraform backend configuration.
func findBackendFilenames(dirs []string, chdir string) (filenames []string, err error) {

	// Start at the highest directory (configuration directory)
	// and go deeper towards the current directory.
	// Files in the current directory take precedence
	// over files in parent directories.
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		// Get a sorted list of files in this directory.
		files, err := getFileNames(dir)
		if err != nil {
			return nil, err
		}
		sort.Strings(files)

		// Add any matching backend files.
		for _, name := range matchFiles(files, "*.tfbackend") {
			filenames = append(filenames, path.Join(dir, name))
		}
	}

	return filenames, nil
}

// loadBackendConfiguration reads *.tfbackend files from the specified directories,
// renders them with Terraform variables, and returns a backend configuration.
func loadBackendConfiguration(dirs []string, chdir string, vars variables) (backend map[string]string, err error) {

	filenames, err := findBackendFilenames(dirs, chdir)
	if err != nil {
		return nil, err
	}

	backend = map[string]string{}

	if len(filenames) == 0 {
		return backend, nil
	}

	for _, filename := range filenames {
		if config, err := parseBackendFile(filename, vars); err != nil {
			return nil, err
		} else {
			for name, value := range config {
				backend[name] = value
			}
		}
	}

	return backend, nil
}

// parseBackendFile parses a *.tfbackend file as HCL into a map of strings.
// Variables can be used in the same way as *.tf files using the `var` object.
func parseBackendFile(filename string, vars variables) (map[string]string, error) {
	// Parse the file.
	p := hclparse.NewParser()
	file, diags := p.ParseHCLFile(filename)
	for _, err := range diags.Errs() {
		return nil, err // TODO: handle multiple errors
	}

	// Decode the file into a map of cty values.
	values := map[string]cty.Value{}
	ctx := varsEvalContext(vars)
	diags = gohcl.DecodeBody(file.Body, ctx, &values)
	for _, err := range diags.Errs() {
		return nil, err // TODO: handle multiple errors
	}

	// Convert the cty values into strings.
	strings := map[string]string{}
	for key, val := range values {
		var s string
		err := gocty.FromCtyValue(val, &s)
		if err != nil {
			return nil, err
		}
		strings[key] = s
	}

	return strings, nil
}

// varsEvalContext returns an EvalContext with a `var` variable
// containing all of the variables.
func varsEvalContext(vars variables) *hcl.EvalContext {
	ctx := hcl.EvalContext{}
	values := map[string]cty.Value{}
	for _, v := range vars {
		values[v.name] = cty.StringVal(v.value)
	}
	ctx.Variables = map[string]cty.Value{"var": cty.ObjectVal(values)}
	return &ctx
}
