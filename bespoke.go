package bespoke

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/matt4biz/envsubst/parse"
)

type Runner struct {
	Args   []string
	Env    []string // key=value pairs, not a map
	Writer io.Writer
	Temp   string
	Debug  bool
}

// Run handles the top-level work: identify the target, create a temp
// directory for the copies, copy/substitute the targeted files, and
// then run kustomize on that temp directory.
func (r *Runner) Run() error {
	cwd, err := os.Getwd()

	if err != nil {
		return fmt.Errorf("can't get CWD: %w", err)
	}

	if len(r.Args) > 0 {
		cwd, err = filepath.Abs(r.Args[0])

		if err != nil {
			return fmt.Errorf("can't fix CWD: %w", err)
		}
	}

	r.Temp, err = os.MkdirTemp(os.TempDir(), "bespoke-")

	if err != nil {
		return fmt.Errorf("can't make tempdir: %w", err)
	}

	if !r.Debug {
		defer r.Cleanup()
	}

	if len(r.Env) == 0 {
		r.Env = os.Environ()
	}

	// recursively read through the kustomization files
	// and find all files that will need to be copied

	target, err := Accumulate(cwd)

	if err != nil {
		return fmt.Errorf("can't process target: %w", err)
	}

	// read from the abs path to a relative path in the
	// temp directory, substituting env vars on the way

	for k, v := range target.top().relativeFiles() {
		fn := r.tempFile(v)
		data, err := r.readFileSkipping(k)

		if err != nil {
			return fmt.Errorf("can't process file: %w", err)
		}

		if err = os.WriteFile(fn, data, 0555); err != nil {
			return fmt.Errorf("can't write file: %w", err)
		}
	}

	if err = r.runKustomize(target); err != nil {
		return fmt.Errorf("can't run tool: %w", err)
	}

	return nil
}

// Accumulate starts in the given directory and walks the tree
// defined by kustomize targets to collect other targets and files
// that we need to copy
func Accumulate(root string) (t *Target, err error) {
	if root, err = filepath.Abs(root); err != nil {
		return nil, err
	}

	fi, err := os.Stat(root)

	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	t = &Target{Root: root}

	return t, t.accumulate()
}

func (r *Runner) Cleanup() {
	if r.Debug {
		_ = os.RemoveAll(r.Temp)
	}
}

// tempFile takes a relative path, returns an absolute path
// within the temp directory, ensuring that any intermediate
// directories have been created.
func (r *Runner) tempFile(rel string) string {
	fn := filepath.Join(r.Temp, rel)
	dir := filepath.Dir(fn)

	if err := os.MkdirAll(dir, 0777); err != nil {
		log.Fatalf("temp dirs: %s", err)
	}

	return fn
}

// readFileSkipping reads a file and substitutes defined env
// vars in it, returning the data to write back out.
func (r *Runner) readFileSkipping(fn string) ([]byte, error) {
	b, err := os.ReadFile(fn)

	if err != nil {
		return nil, err
	}

	restrict := parse.Restrictions{NoFail: true}
	s, err := parse.New("file", r.Env, &restrict).Parse(string(b))

	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

// runKustomize actually uses the kustomize API directly.
// TODO - add support for the build flags the real too accepts
func (r *Runner) runKustomize(target *Target) error {
	root, err := filepath.Rel(target.top().Root, target.Root)

	if err != nil {
		return fmt.Errorf("can't relativize: %w", err)
	}

	fSys := filesys.MakeFsOnDisk()
	pc := types.EnabledPluginConfig(types.BploUseStaticallyLinked)

	pc.HelmConfig.Enabled = true
	pc.HelmConfig.Command = "helm"

	opts := krusty.Options{PluginConfig: pc}
	k := krusty.MakeKustomizer(&opts)
	m, err := k.Run(fSys, filepath.Join(r.Temp, root))

	if err != nil {
		return err
	}

	yml, err := m.AsYaml()

	if err != nil {
		return err
	}

	_, err = r.Writer.Write(yml)

	return err
}
