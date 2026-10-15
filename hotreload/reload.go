package hotreload

import (
	"sync"

	"github.com/23dd/animation/engine"
	"github.com/23dd/animation/script"
)

type Reloader struct {
	Path          string
	Width, Height int
	Engine        *engine.Engine
	mu            sync.Mutex
}

func (r *Reloader) Reload() {
	r.mu.Lock()
	defer r.mu.Unlock()
	runtime, err := script.LoadFile(r.Path, r.Width, r.Height)
	if err != nil {
		r.Engine.Swap(nil, nil, err)
		return
	}
	r.Engine.Swap(runtime.Scene, runtime.Timeline, nil)
}
