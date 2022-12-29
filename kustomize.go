package bespoke

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
)

func (t *Target) readKustomization() (*types.Kustomization, error) {
	var count int
	var kfile types.Kustomization

	for _, file := range konfig.RecognizedKustomizationFileNames() {
		path := filepath.Join(t.Root, file)
		data, err := os.ReadFile(path)

		if err != nil {
			continue
		}

		count++

		// we always must copy this k-file along with the rest
		t.Files = append(t.Files, file)

		if err = kfile.Unmarshal(data); err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("%d files read", count)
	}

	return &kfile, nil
}

func (t *Target) accumulate() error { //nolint:gocyclo
	k, err := t.readKustomization()

	if err != nil {
		return err
	}

	for _, b := range k.Bases {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.Resources {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.Components {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.Crds {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.Configurations {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.Validators {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.Transformers {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.Generators {
		if err = t.accumulateEntry(b); err != nil {
			return err
		}
	}

	for _, b := range k.PatchesStrategicMerge {
		if err = t.accumulateEntry(string(b)); err != nil {
			return err
		}
	}

	for _, b := range k.Patches {
		if err = t.accumulateEntry(b.Path); err != nil {
			return err
		}
	}

	for _, b := range k.PatchesJson6902 {
		if err = t.accumulateEntry(b.Path); err != nil {
			return err
		}
	}

	for _, b := range k.ConfigMapGenerator {
		for _, f := range b.FileSources {
			if err = t.accumulateEntry(f); err != nil {
				return err
			}
		}
	}

	for _, b := range k.SecretGenerator {
		for _, f := range b.FileSources {
			if err = t.accumulateEntry(f); err != nil {
				return err
			}
		}
	}

	for _, b := range k.HelmCharts {
		if err = t.accumulateEntry(b.Repo); err != nil {
			return err
		}

		if err = t.accumulateEntry(b.ValuesFile); err != nil {
			return err
		}
	}

	return nil
}

func (t *Target) accumulateEntry(path string) error {
	root := path

	// if it's a web resource, ignore it
	// if it's a dir, run accumulate on the target from the path
	// if it's a file, add it to the current target's file list

	if strings.HasPrefix(path, "http") {
		return nil
	}

	if !filepath.IsAbs(path) {
		root = filepath.Join(t.Root, path)
	}

	fi, err := os.Stat(root)

	if err != nil {
		// not something on disk, so ignore it
		return nil
	}

	if fi.IsDir() {
		t1, err := t.pathToTarget(path)

		if err != nil {
			return err
		}

		return t1.accumulate()
	}

	if fi.Mode().IsRegular() {
		t.Files = append(t.Files, path)
		return nil
	}

	return fmt.Errorf("cannot accumulate %s: %s", root, fi.Mode())
}
