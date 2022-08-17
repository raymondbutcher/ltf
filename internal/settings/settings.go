package settings

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/raymondbutcher/ltf/internal/hook"
	"gopkg.in/yaml.v2"
)

type settings struct {
	Hooks hook.Hooks `yaml:"hooks"`
}

func Load(cwd string) (*settings, error) {
	file, err := findFile(cwd)
	if err != nil {
		return nil, err
	}

	settings := settings{}

	if file == "" {
		return &settings, nil
	}

	rel, err := filepath.Rel(cwd, file)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(rel)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict([]byte(content), &settings); err != nil {
		return nil, err
	}

	for name, hook := range settings.Hooks {
		hook.Name = name
	}

	return &settings, nil
}

// findFile returns the path to ltf.yaml in the current or parent directories,
// or an empty string if not found.
func findFile(dir string) (string, error) {
	lastDir := ""
	for {
		// Check this directory.
		f, err := os.Open(dir)
		if err != nil {
			return "", err
		}
		names, err := f.Readdirnames(0)
		if err != nil {
			return "", err
		}
		for _, name := range names {
			if name == "ltf.yaml" {
				return path.Join(dir, name), nil
			}
		}

		// Move to the parent directory.
		dir = path.Dir(dir)

		// Stop if this directory was already checked.
		// This occurs after reaching the filesystem root.
		if dir == lastDir {
			break
		}
		lastDir = dir
	}

	// Not found.
	return "", nil
}
