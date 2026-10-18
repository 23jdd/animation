package app

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativePPTXFlagsAndCommand(t *testing.T) {
	for _, args := range [][]string{{"-native", "-time", "0"}, {"-native", "-times", "0,1"}, {"-native", "-interval", "1"}} {
		if _, _, err := flags("pptx", args); err == nil {
			t.Errorf("accepted conflicting native/frame options %v", args)
		}
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "native.js")
	if err := os.WriteFile(source, []byte(`const c=circle({radius:20});add(c);play(moveTo(c,[100,0]),{duration:2,ease:"smooth"});`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"ppt", "pptx"} {
		output := filepath.Join(dir, command+".pptx")
		if err := Run([]string{command, "-native", "-fps", "30", source, output}, io.Discard); err != nil {
			t.Fatal(err)
		}
		z, err := zip.OpenReader(output)
		if err != nil {
			t.Fatal(err)
		}
		for _, part := range z.File {
			if strings.HasPrefix(part.Name, "ppt/media/") {
				t.Fatal("-native command exported screenshots")
			}
		}
		z.Close()
	}
}
