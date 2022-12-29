package bespoke

import (
	"fmt"
	"path/filepath"
	"strings"
)

const upOne = ".."

type Target struct {
	Parent   *Target
	Root     string
	Files    []string
	Children map[string]*Target
}

// pathToTarget takes a relative path which may include '..' and
// finds/creates a target above or below (and intervening targets
// as needed) and returns the one to work on
func (t *Target) pathToTarget(path string) (*Target, error) {
	var err error

	if !filepath.IsAbs(path) {
		path = filepath.Clean(filepath.Join(t.Root, path))
	}

	relative, err := filepath.Rel(t.Root, path)

	if err != nil {
		return nil, err
	}

	if relative == "." {
		return t, nil
	}

	if strings.HasPrefix(relative, upOne) {
		pd := filepath.Dir(t.Root)
		rem, err := filepath.Rel(upOne, relative)

		if err != nil {
			return nil, err
		}

		if t.Parent != nil {
			if t.Parent.Root == pd {
				return t.Parent.pathToTarget(rem)
			}

			return nil, fmt.Errorf("invalid parent %s", pd)
		}

		parent := &Target{Root: pd, Children: map[string]*Target{filepath.Base(t.Root): t}}

		t.Parent = parent

		return parent.pathToTarget(rem)
	}

	parts := strings.SplitN(relative, string(filepath.Separator), 2)

	if child, ok := t.Children[parts[0]]; ok {
		if len(parts) == 1 || len(parts[1]) == 0 {
			return child, nil
		}

		return child.pathToTarget(parts[1])
	}

	child := &Target{Root: filepath.Join(t.Root, parts[0]), Parent: t}

	if t.Children == nil {
		t.Children = make(map[string]*Target)
	}

	t.Children[parts[0]] = child

	if len(parts) == 1 || len(parts[1]) == 0 {
		return child, nil
	}

	return child.pathToTarget(parts[1])
}

// top returns the top-most directory, i.e., the root path
// of the "oldest" ancestor above this target, which will
// be the path to use to make relative all files to create
func (t *Target) top() *Target {
	if t.Parent == nil {
		return t
	}

	return t.Parent.top()
}

// files returns a list of absolute file paths from this
// target and all its children; so it should be run on a
// "top" target to get the whole tree
func (t *Target) files() []string {
	if len(t.Files) == 0 && len(t.Children) == 0 {
		return nil
	}

	var result []string

	for _, f := range t.Files {
		result = append(result, filepath.Join(t.Root, f))
	}

	for _, c := range t.Children {
		result = append(result, c.files()...)
	}

	return result
}

// relativeFiles returns a map of absolute file paths from this
// target and all its children to their relative paths from this
// target's root; so it should be run on a "top" target  to get
// the whole tree
func (t *Target) relativeFiles() map[string]string {
	if len(t.Files) == 0 && len(t.Children) == 0 {
		return nil
	}

	result := make(map[string]string)

	for _, f := range t.files() {
		if r, err := filepath.Rel(t.Root, f); err == nil { // inverted
			result[f] = r
		}
	}

	return result
}
