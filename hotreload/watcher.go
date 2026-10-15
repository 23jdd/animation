package hotreload

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
	done    chan struct{}
	once    sync.Once
}

func New(path string, debounce time.Duration, onChange func()) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		w.Close()
		return nil, err
	}
	root := filepath.Dir(abs)
	if err = addDirectoryTree(w, root); err != nil {
		w.Close()
		return nil, err
	}
	out := &Watcher{watcher: w, done: make(chan struct{})}
	go func() {
		var timer *time.Timer
		for {
			select {
			case e, ok := <-w.Events:
				if !ok {
					return
				}
				clean := filepath.Clean(e.Name)
				rel, relErr := filepath.Rel(root, clean)
				if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					continue
				}
				if e.Op&fsnotify.Create != 0 {
					if info, statErr := os.Stat(clean); statErr == nil && info.IsDir() {
						_ = addDirectoryTree(w, clean)
						continue
					}
				}
				ext := strings.ToLower(filepath.Ext(e.Name))
				if ext != ".js" && ext != ".json" && ext != ".svg" && ext != ".ttf" && ext != ".otf" && ext != ".png" && ext != ".jpg" {
					continue
				}
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounce, onChange)
			case <-w.Errors:
			case <-out.done:
				return
			}
		}
	}()
	return out, nil
}

func addDirectoryTree(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if path != root && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		return w.Add(path)
	})
}
func (w *Watcher) Close() error {
	var err error
	w.once.Do(func() { close(w.done); err = w.watcher.Close() })
	return err
}
