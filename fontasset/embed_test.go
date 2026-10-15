package fontasset

import (
	"strings"
	"testing"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

func TestEmbeddedJetBrainsMono(t *testing.T) {
	assertFontContains(t, "JetBrains Mono", JetBrainsMonoRegular(), "Animation Engine")
	if !strings.Contains(strings.ToUpper(JetBrainsMonoLicense()), "SIL OPEN FONT LICENSE") {
		t.Fatal("embedded JetBrains Mono license is missing")
	}
}

func TestEmbeddedSourceHanSansSC(t *testing.T) {
	assertFontContains(t, "Source Han Sans SC", SourceHanSansSCRegular(), "Animation动画播放")
	if !strings.Contains(strings.ToUpper(SourceHanSansLicense()), "SIL OPEN FONT LICENSE") {
		t.Fatal("embedded Source Han Sans license is missing")
	}
}

func assertFontContains(t *testing.T, name string, data []byte, characters string) {
	t.Helper()
	parsed, err := opentype.Parse(data)
	if err != nil {
		t.Fatalf("parse embedded %s: %v", name, err)
	}
	var buffer sfnt.Buffer
	for _, character := range characters {
		if index, err := parsed.GlyphIndex(&buffer, character); err != nil || index == 0 {
			t.Fatalf("embedded %s is missing %q", name, character)
		}
	}
}
