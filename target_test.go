package bespoke

import (
	"reflect"
	"sort"
	"testing"
)

func TestTargetsGrandparent(t *testing.T) {
	target := Target{Root: "/a/b/c"}
	t1, err := target.pathToTarget("../..")

	if err != nil {
		t.Errorf("failed: %s", err)
	}

	if target.Parent == nil {
		t.Fatalf("no parent created")
	}

	if target.Parent.Root != "/a/b" {
		t.Errorf("invalid parent: %s", target.Parent.Root)
	}

	if pd := target.Parent.Children["c"]; pd != &target {
		t.Errorf("invalid parent's children: %v", target.Children)
	}

	if t1.Root != "/a" {
		t.Errorf("invalid target: %s", t1.Root)
	}

	if cd := t1.Children["b"]; cd != target.Parent {
		t.Errorf("invalid target's children: %v", t1.Children)
	}

	if t2 := target.top(); t2.Root != "/a" {
		t.Errorf("invalid top directory: %v", t2)
	}
}

func TestTargetsGrandchild(t *testing.T) {
	target := Target{Root: "/a"}
	t1, err := target.pathToTarget("b/c")

	if err != nil {
		t.Errorf("failed: %s", err)
	}

	if t1.Root != "/a/b/c" {
		t.Errorf("invalid target: %s", t1.Root)
	}

	if t1.Parent == nil {
		t.Fatalf("target has no parent")
	}

	if t1.Parent.Root != "/a/b" {
		t.Errorf("invalid parent path: %s", t1.Parent.Root)
	}

	if cd := target.Children["b"]; cd != t1.Parent {
		t.Errorf("invalid target's children: %v", target)
	}

	if cd := t1.Parent.Children["c"]; cd != t1 {
		t.Errorf("invalid intermediate children: %v", t1.Parent.Children)
	}

	if t1.Parent.Parent != &target {
		t.Errorf("invalid intermediate parent: %v", t1.Parent)
	}

	if t2 := target.top(); t2.Root != "/a" {
		t.Errorf("invalid top directory: %v", t2)
	}
}

func TestTargetsGrandNephew(t *testing.T) {
	target := Target{Root: "/a/b/c"}
	t1, err := target.pathToTarget("../../d")

	if err != nil {
		t.Errorf("failed: %s", err)
	}

	if t1.Root != "/a/d" {
		t.Errorf("invalid target: %s", t1.Root)
	}

	if t1.Parent == nil {
		t.Fatalf("target's parent missing")
	}

	if target.Parent == nil {
		t.Fatalf("no parent created")
	}

	if target.Parent.Root != "/a/b" {
		t.Errorf("invalid parent: %s", target.Parent.Root)
	}

	if pd := target.Parent.Children["c"]; pd != &target {
		t.Errorf("invalid parent's children: %v", target.Children)
	}

	if t1.Parent != target.Parent.Parent {
		t.Errorf("invalid common ancestor %v & %v", t1, target.Parent)
	}

	if target.Parent.Parent.Root != "/a" {
		t.Errorf("invalid common ancestor root: %s", target.Parent.Parent.Root)
	}

	if cd := target.Parent.Parent.Children["d"]; cd != t1 {
		t.Errorf("invalid common ancestor children: %v", target.Parent.Parent.Children)
	}
}

func TestTargetFiles(t *testing.T) {
	expected := []string{"/a/b/b1", "/a/b/c/c1", "/a/b/c/c2", "/a/d/d1"}
	relative := map[string]string{"/a/b/b1": "b/b1", "/a/b/c/c1": "b/c/c1", "/a/b/c/c2": "b/c/c2", "/a/d/d1": "d/d1"}

	target := Target{Root: "/a/b/c", Files: []string{"c1", "c2"}}
	parent := Target{Root: "/a/b", Files: []string{"b1"}, Children: map[string]*Target{"c": &target}}
	lateral := Target{Root: "/a/d", Files: []string{"d1"}}
	grandpapa := Target{Root: "/a", Children: map[string]*Target{"b": &parent, "d": &lateral}}

	target.Parent = &parent
	parent.Parent = &grandpapa
	lateral.Parent = &grandpapa

	t1 := target.top()
	files := t1.files()

	sort.Strings(files)

	if !reflect.DeepEqual(files, expected) {
		t.Errorf("invalid file list %s", files)
	}

	rel := t1.relativeFiles()

	if !reflect.DeepEqual(rel, relative) {
		t.Errorf("invalid relative list %s", rel)
	}
}
