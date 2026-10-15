package engine

type State struct {
	Playing               bool
	Time, Duration, Speed float64
	Error                 error
}

func (e *Engine) State() State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return State{Playing: e.Playing, Time: e.last, Duration: e.Timeline.Duration(), Speed: e.Speed, Error: e.ScriptError}
}
