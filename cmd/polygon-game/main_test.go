package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/23dd/animation/polygon"
)

func TestCommandGeneratesGoResultsAndJavaScript(t *testing.T) {
	dir := t.TempDir()
	jsPath, jsonPath, reportPath := filepath.Join(dir, "animation.js"), filepath.Join(dir, "result.json"), filepath.Join(dir, "result.txt")
	var stdout, stderr bytes.Buffer
	err := run([]string{"-input", "-", "-js", jsPath, "-json", jsonPath, "-report", reportPath}, strings.NewReader("4\n-7 4 2 5\n+ * * +\n"), &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "最大得分: 33") {
		t.Fatal(stdout.String())
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var result polygon.Result
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.MaxScore != "33" {
		t.Fatal(result.MaxScore)
	}
	js, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(js, []byte("controller.update")) || !bytes.Contains(js, []byte(`"max_score": "33"`)) {
		t.Fatal("missing JS scene or Go trace")
	}
	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(report, []byte("[1 4]")) {
		t.Fatal(string(report))
	}
}

func TestCommandRejectsInvalidOptions(t *testing.T) {
	for _, args := range [][]string{{"-fps", "NaN"}, {"-fps", "101"}, {"-scale", "0"}, {"-time", "-2"}, {"-gif", "bad.png"}, {"-js", "same.js", "-report", "same.js"}, {"unexpected"}} {
		var out bytes.Buffer
		if err := run(args, strings.NewReader(""), &out, &out); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
