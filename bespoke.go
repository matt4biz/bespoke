package bespoke

import (
	"fmt"
	"os"
	"path/filepath"
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

func Run(args []string) int {
	fmt.Println("run bespoke:", args)

	cwd, err := os.Getwd()

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get CWD: %s\n", err)
		return 1
	}

	if len(args) > 0 {
		cwd, err = filepath.Abs(args[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "can't fix CWD: %s\n", err)
			return 1
		}
	}

	fmt.Println(cwd)

	dir, err := os.MkdirTemp(os.TempDir(), "bespoke-")

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't make tempdir: %s\n", err)
		return 1
	}

	fmt.Println(dir)

	defer os.RemoveAll(dir)

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
