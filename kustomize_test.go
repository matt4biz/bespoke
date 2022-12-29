package bespoke

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAccumulate(t *testing.T) {
	target, err := Accumulate("samples/files/overlays/dev")

	if err != nil {
		t.Fatal(err)
	}

	rel := target.top().relativeFiles()
	files := make([]string, 0, len(rel))

	for _, v := range rel {
		files = append(files, v)
	}

	sort.Strings(files)

	expected := []string{
		"bases/deployment.yaml",
		"bases/kustomization.yaml",
		"bases/routing.yaml",
		"overlays/dev/kustomization.yaml",
		"overlays/dev/patch.yaml",
	}

	if !reflect.DeepEqual(expected, files) {
		t.Errorf("invalid list %s", cmp.Diff(expected, files))
	}
}
