package scene

import (
	"github.com/23dd/animation/mobject"
	"sync"
)

type Scene struct {
	mu      sync.RWMutex
	Objects []mobject.Mobject
	Camera  *Camera
}

func New(width, height int) *Scene { return &Scene{Camera: NewCamera(width, height)} }
func (s *Scene) Add(objects ...mobject.Mobject) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Objects = append(s.Objects, objects...)
}
func (s *Scene) Remove(target mobject.Mobject) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, o := range s.Objects {
		if o.ID() == target.ID() {
			s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
			return
		}
	}
}
func (s *Scene) Snapshot() []mobject.Mobject {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]mobject.Mobject(nil), s.Objects...)
}
func (s *Scene) Update(time, dt float64) error {
	for _, o := range s.Snapshot() {
		if err := o.Update(time, dt); err != nil {
			return err
		}
	}
	return nil
}
