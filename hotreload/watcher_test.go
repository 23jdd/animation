package hotreload

import "testing"

func TestWatchableImageExtensions(t *testing.T) {
	for _, path := range []string{"photo.png", "assets/photo.JPG", "photo.jpeg", "PHOTO.JPEG", "photo.webp", "photo.GIF", "scene.js", "font.ttf"} {
		if !watchableFile(path) {
			t.Errorf("supported asset not watched: %s", path)
		}
	}
	for _, path := range []string{"README.md", "output.mp4", "backup.png.tmp", "photo"} {
		if watchableFile(path) {
			t.Errorf("unexpected watched file: %s", path)
		}
	}
}
