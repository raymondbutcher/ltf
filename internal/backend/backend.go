package backend

import (
	"fmt"
	"path"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/raymondbutcher/ltf"
	"github.com/raymondbutcher/ltf/internal/filesystem"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

// LoadConfiguration reads *.tfbackend files from the specified directories,
// renders them with Terraform variables, and returns a backend configuration.
func LoadConfiguration(dirs []string, chdir string, vars ltf.VariableService) (map[string]string, error) {
	filenames, err := findBackendFiles(dirs, chdir)
	if err != nil {
		return nil, err
	}

	backend := map[string]string{}

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

// findBackendFiles returns *.tfbackend files to use for the Terraform backend configuration.
func findBackendFiles(dirs []string, chdir string) (filenames []string, err error) {
	// Start at the highest directory (configuration directory)
	// and go deeper towards the current directory.
	// Files in the current directory take precedence
	// over files in parent directories.
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		// Get a sorted list of files in this directory.
		files, err := filesystem.ReadNames(dir)
		if err != nil {
			return nil, err
		}
		sort.Strings(files)

		// Add any matching backend files.
		for _, name := range filesystem.MatchNames(files, "*.tfbackend") {
			filenames = append(filenames, path.Join(dir, name))
		}
	}

	return filenames, nil
}

// parseBackendFile parses a *.tfbackend file as HCL into a map of strings.
// Variables can be used in the same way as *.tf files using the `var` object.
func parseBackendFile(filename string, vars ltf.VariableService) (map[string]string, error) {
	// Parse the file.
	p := hclparse.NewParser()
	file, diags := p.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing %s: %s", filename, diags.Error())
	}

	// Decode the file into a map of cty values.
	values := map[string]cty.Value{}
	ctx, err := varEvalContext(vars)
	if err != nil {
		return nil, fmt.Errorf("creating backend context for %s: %w", filename, err)
	}
	diags = gohcl.DecodeBody(file.Body, ctx, &values)
	if diags.HasErrors() {
		return nil, fmt.Errorf("decoding hcl %s: %s", filename, diags.Error())
	}

	// Convert the cty values into strings.
	strings := map[string]string{}
	for key, val := range values {
		if val.Type() == cty.String {
			var s string
			err := gocty.FromCtyValue(val, &s)
			if err != nil {
				return nil, fmt.Errorf("converting string value in %s: %w", filename, err)
			}
			strings[key] = s
		} else {
			b, err := json.Marshal(val, val.Type())
			if err != nil {
				return nil, fmt.Errorf("converting non-string value in %s: %w", filename, err)
			}
			strings[key] = string(b)
		}
	}

	return strings, nil
}

// varEvalContext returns an EvalContext with a `var` object containing variables.
func varEvalContext(vars ltf.VariableService) (*hcl.EvalContext, error) {
	values := map[string]cty.Value{}
	for _, v := range vars.Each() {
		ct, err := gocty.ImpliedType(v.AnyValue)
		if err != nil {
			return nil, fmt.Errorf("getting cty type for var.%s (%v): %w", v.Name, v.AnyValue, err)
		}
		cv, err := gocty.ToCtyValue(v.AnyValue, ct)
		if err != nil {
			return nil, fmt.Errorf("converting to cty type: %w", err)
		}
		values[v.Name] = cv
	}
	ctx := hcl.EvalContext{}
	ctx.Variables = map[string]cty.Value{
		"var": cty.ObjectVal(values),
	}
	return &ctx, nil
}
