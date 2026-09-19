//go:build unix

package term

import (
	"os"
	"os/signal"
	"syscall"
)

// resizeLoop ждёт SIGWINCH и рассылает ResizeEvent при изменении размера.
func (rt *rawTerminal) resizeLoop() {
	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	defer signal.Stop(sigwinch)

	rt.lastW, rt.lastH = rt.Size()

	for {
		select {
		case <-rt.stopCh:
			return
		case <-sigwinch:
			rt.handleResize()
		}
	}
}

// handleResize читает новый размер, сравнивает с предыдущим
// и рассылает событие + вызывает обработчики.
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
