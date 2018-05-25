package tools_test

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/arsham/blush/internal/tools"
)

func stringSliceEq(a, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func setup(count int) (dirs, expect []string, cleanup func(), err error) {
	ret := make(map[string]struct{})
	tmp, err := ioutil.TempDir("", "blush")
	if err != nil {
		return nil, nil, func() {}, err
	}
	cleanup = func() {
		os.RemoveAll(tmp)
	}
	files := []struct {
		dir   string
		count int
	}{
		{"a", count},     // keep this here.
		{"a/b/c", count}, // this one is in the above folder, keep!
		{"abc", count},   // this one is outside.
		{"f", 0},         // this should not be matched.
	}
	for _, f := range files {
		l := path.Join(tmp, f.dir)
		err := os.MkdirAll(l, os.ModePerm)
		if err != nil {
			return nil, nil, cleanup, err
		}

		for i := 0; i < f.count; i++ {
			f, err := ioutil.TempFile(l, "file_")
			if err != nil {
				return nil, nil, cleanup, err
			}
			ret[path.Dir(f.Name())] = struct{}{}
			expect = append(expect, f.Name())
		}
	}
	for d := range ret {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return
}

func TestFilesError(t *testing.T) {
	f, err := tools.Files(false)
	if f != nil {
		t.Errorf("f = %v, want nil", f)
	}
	if err == nil {
		t.Error("err = nil, want error")
	}
	f, err = tools.Files(false, "/path to heaven")
	if err == nil {
		t.Error("err = nil, want error")
	}
	if f != nil {
		t.Errorf("f = %v, want nil", f)
	}
}

func TestFiles(t *testing.T) {
	dirs, expect, cleanup, err := setup(10)
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}

	f, err := tools.Files(false, dirs...)
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if f == nil {
		t.Error("f = nil, want []string")
	}
	if !stringSliceEq(expect, f) {
		t.Errorf("f = %v, \nwant %v", f, expect)
	}

	// the a and abc should match, a/b/c should not
	f, err = tools.Files(false, dirs[0], dirs[2])
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if len(f) != 20 { // all files in `a` and `abc`
		t.Errorf("len(f) = %d, want %d: %v", len(f), 20, f)
	}
}

func TestFilesOnSingleFile(t *testing.T) {
	file, err := ioutil.TempFile("", "blush_tools")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()
	defer func() {
		if err = os.Remove(name); err != nil {
			t.Error(err)
		}
	}()

	f, err := tools.Files(true, name)
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if len(f) != 1 {
		t.Fatalf("len(f) = %d, want 1", len(f))
	}
	if f[0] != name {
		t.Errorf("f[0] = %s, want %s", f[0], name)
	}

	f, err = tools.Files(false, name)
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if len(f) != 1 {
		t.Fatalf("len(f) = %d, want 1", len(f))
	}
	if f[0] != name {
		t.Errorf("f[0] = %s, want %s", f[0], name)
	}
}

func TestFilesRecursive(t *testing.T) {
	f, err := tools.Files(true, "/path to heaven")
	if err == nil {
		t.Error("err = nil, want error")
	}
	if f != nil {
		t.Errorf("f = %v, want nil", f)
	}

	dirs, expect, cleanup, err := setup(10)
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}

	f, err = tools.Files(true, dirs...)
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if f == nil {
		t.Error("f = nil, want []string")
	}
	if !stringSliceEq(expect, f) {
		t.Errorf("f = %v, want %v", f, expect)
	}

	f, err = tools.Files(true, dirs[0]) // expecting `a`
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if len(f) != 20 { // all files in `a`
		t.Errorf("len(f) = %d, want %d: %v", len(f), 20, f)
	}
}
