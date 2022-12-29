package bespoke

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/matt4biz/envsubst/parse"
)

// we get Root, we make enclosingDir, which contains all the dirs including the copy of root
// as we process dirs from Root's k-file, we will need to make them under enclosingDir
//
// - top
//   |
//   - bases
//     |
//     - apps
//       |
//       - deployment.yaml
//       - kustomization.yaml, which references deployment.yaml
//     - kustomization.yaml, which references apps
//   - overlays
//     |
//     - dev
//       |
//       - kustomization.yaml which references ../../bases
//     - prd
//
// we'll make ED and within it bases, bases/apps, overlays, overlays/dev
// but we start with top/overlays/dev/
//
// so each time we start a new accum dir target, we must make the dir
// we may be called within that dev dir, so we don't originally know
// what the top-level dir is going to be
//
// so we should accumulate files/dirs first, and then do the copying
// once we know what dirs to make / where they go ...

// 1. create a tempdir
// 2. run the add-target loop over the input dir
// 3. as files are copied, substitute env vars
// 4. then run kustomize on the tempdir

type Runner struct {
	Args   []string
	Env    []string // key=value pairs, not a map
	Writer io.Writer
	Temp   string
	Debug  bool
}

func (r *Runner) Run() int {
	cwd, err := os.Getwd()

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get CWD: %s\n", err)
		return 1
	}

	if len(r.Args) > 0 {
		cwd, err = filepath.Abs(r.Args[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "can't fix CWD: %s\n", err)
			return 1
		}
	}

	r.Temp, err = os.MkdirTemp(os.TempDir(), "bespoke-")

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't make tempdir: %s\n", err)
		return 1
	}

	if !r.Debug {
		defer r.Cleanup()
	}

	if len(r.Env) == 0 {
		r.Env = os.Environ()
	}

	target, err := Accumulate(cwd)

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't process target: %s\n", err)
		return 1
	}

	for k, v := range target.top().relativeFiles() {
		fn := r.tempFile(v)
		data, err := r.readFileSkipping(k)

		if err != nil {
			fmt.Fprintf(os.Stderr, "can't process file: %s\n", err)
			return 1
		}

		if err = os.WriteFile(fn, data, 0555); err != nil {
			fmt.Fprintf(os.Stderr, "can't write file: %s\n", err)
			return 1
		}
	}

	root, err := filepath.Rel(target.top().Root, target.Root)

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't relativize: %s\n", err)
		return 1
	}

	if err = r.runKustomize(root); err != nil {
		fmt.Fprintf(os.Stderr, "can't run tool: %s\n", err)
		return 1
	}

	return 0
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

func (r *Runner) tempFile(rel string) string {
	fn := filepath.Join(r.Temp, rel)
	dir := filepath.Dir(fn)

	_ = os.MkdirAll(dir, 0777)

	return fn
}

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

func (r *Runner) runKustomize(root string) error {
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
