package bespoke

import (
	"bytes"
	"os"
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

func TestKustomize(t *testing.T) {
	buffer := bytes.Buffer{}
	env := []string{"LC_NGINX_VERSION=1.14.2", "LC_APP=httpbin", "LC_PORT=8000"}
	runner := Runner{Args: []string{"samples/files/overlays/dev"}, Env: env, Writer: &buffer}
	code := runner.Run()

	if code != 0 {
		t.Fatalf("run failed: %d", code)
	}

	golden, err := os.ReadFile("samples/golden.yaml")

	if err != nil {
		t.Fatalf("can't read golden data: %s", err)
	}

	if !cmp.Equal(golden, buffer.Bytes()) {
		t.Errorf("invalid output: %s", cmp.Diff(string(golden), buffer.String()))
	}
}
