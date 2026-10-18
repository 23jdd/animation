package export

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestPPTXPackageXMLAndRelationships(t *testing.T) {
	s, tl, r := pptxTestScene()
	file := filepath.Join(t.TempDir(), "玫瑰 & 分镜.pptx")
	if err := PPTX(file, s, tl, r, []float64{2, 0, 1, 1}); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	parts := make(map[string][]byte)
	for _, f := range z.File {
		stream, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			t.Fatal(err)
		}
		if _, exists := parts[f.Name]; exists {
			t.Fatalf("duplicate ZIP entry %s", f.Name)
		}
		parts[f.Name] = data
	}
	for name, data := range parts {
		if strings.HasSuffix(name, ".xml") || strings.HasSuffix(name, ".rels") {
			decoder := xml.NewDecoder(bytes.NewReader(data))
			for {
				_, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("invalid XML in %s: %v", name, err)
				}
			}
		}
		if strings.HasSuffix(name, ".rels") {
			var rels pptxRelationships
			if err := xml.Unmarshal(data, &rels); err != nil {
				t.Fatal(err)
			}
			seen := make(map[string]bool)
			for _, rel := range rels.Items {
				if rel.ID == "" || seen[rel.ID] {
					t.Fatalf("duplicate/empty relationship ID in %s", name)
				}
				seen[rel.ID] = true
				target := path.Clean(path.Join(path.Dir(path.Dir(name)), rel.Target))
				if parts[target] == nil {
					t.Fatalf("dangling relationship %s -> %s", name, target)
				}
			}
		}
	}
	for _, required := range []string{"[Content_Types].xml", "ppt/presProps.xml", "ppt/theme/theme1.xml", "ppt/slideMasters/slideMaster1.xml", "ppt/slideLayouts/slideLayout1.xml", "ppt/notesMasters/notesMaster1.xml"} {
		if parts[required] == nil {
			t.Fatalf("missing required package part %s", required)
		}
	}
	// PowerPoint requires a distinct theme part for the notes master. A shared
	// theme passes the XML schema but makes desktop PowerPoint reject the file.
	if parts["ppt/theme/theme2.xml"] == nil || bytes.Contains(parts["ppt/notesMasters/_rels/notesMaster1.xml.rels"], []byte("theme1.xml")) {
		t.Fatal("notes and slide masters must not share a theme part")
	}
	var core struct {
		Title string `xml:"title"`
	}
	if err := xml.Unmarshal(parts["docProps/core.xml"], &core); err != nil || core.Title != "玫瑰 & 分镜" {
		t.Fatalf("Unicode/XML-special metadata was corrupted: %q %v", core.Title, err)
	}
	var pres struct {
		Slides []struct{} `xml:"sldIdLst>sldId"`
		Size   struct {
			W int `xml:"cx,attr"`
			H int `xml:"cy,attr"`
		} `xml:"sldSz"`
	}
	if err := xml.Unmarshal(parts["ppt/presentation.xml"], &pres); err != nil || len(pres.Slides) != 4 {
		t.Fatalf("incorrect slide list: %v %v", pres.Slides, err)
	}
	for i, want := range []string{"2", "0", "1", "1"} {
		note := parts[fmt.Sprintf("ppt/notesSlides/notesSlide%d.xml", i+1)]
		if !bytes.Contains(note, []byte("Animation time: "+want+" seconds")) {
			t.Fatalf("slide %d lost its timestamp or selection order", i+1)
		}
	}
	if pres.Size.W != 12192000 || pres.Size.H != 6858000 {
		t.Fatalf("incorrect slide dimensions: %+v", pres.Size)
	}
}

func TestPPTXOutputIsReproducible(t *testing.T) {
	s, tl, r := pptxTestScene()
	file := filepath.Join(t.TempDir(), "deck.pptx")
	if err := PPTX(file, s, tl, r, nil); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := PPTX(file, s, tl, r, nil); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(file)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("identical scenes and frame selections should produce identical PPTX bytes")
	}
}
