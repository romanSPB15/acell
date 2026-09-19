//go:build windows

package term

import "time"

// resizeLoop опрашивает размер терминала каждые 100 мс.
func (rt *rawTerminal) resizeLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	rt.lastW, rt.lastH = rt.Size()

	for {
		select {
		case <-rt.stopCh:
			return
		case <-ticker.C:
			rt.handleResize()
		}
	}
}

func (rt *rawTerminal) handleResize() {
	nw, nh := rt.Size()
	if nw == 0 && nh == 0 {
		return
	}
	if nw == rt.lastW && nh == rt.lastH {
		return
	}
	rt.lastW, rt.lastH = nw, nh
	rt.emit(&ResizeEvent{Width: nw, Height: nh})

	rt.resizeMu.Lock()
	handlers := rt.resizeHandlers
	rt.resizeMu.Unlock()
	for _, fn := range handlers {
		fn(nw, nh)
	}
}
