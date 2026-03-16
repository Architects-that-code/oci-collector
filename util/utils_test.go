package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToJSONAndToYAML(t *testing.T) {
	type sample struct {
		Name string `json:"name" yaml:"name"`
		Num  int    `json:"num" yaml:"num"`
	}
	in := sample{Name: "alice", Num: 7}

	j, err := ToJSON(in)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	sj := string(j)
	if !strings.Contains(sj, `"name":"alice"`) || !strings.Contains(sj, `"num":7`) {
		t.Fatalf("unexpected json payload: %s", sj)
	}

	y, err := ToYAML(in)
	if err != nil {
		t.Fatalf("ToYAML error: %v", err)
	}
	sy := string(y)
	if !strings.Contains(sy, "name: alice") || !strings.Contains(sy, "num: 7") {
		t.Fatalf("unexpected yaml payload: %s", sy)
	}
}

func TestWriteToFileWritesUnderHomeDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	const fileName = "collector-test-output.txt"
	const content = "hello world"

	if err := WriteToFile(fileName, []byte(content)); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	fullPath := filepath.Join(home, fileName)
	b, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read output file error: %v", err)
	}
	if string(b) != content {
		t.Fatalf("unexpected file content: got %q, want %q", string(b), content)
	}
}
