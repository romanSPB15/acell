package term

// WindowFocusEvent — окно терминала получило или потеряло фокус
// (DECSET 1004, CSI I / CSI O).
type WindowFocusEvent struct {
	Focused bool
}

const (
	focusInSeq  = "\x1b[I"
	focusOutSeq = "\x1b[O"
)

func parseFocusEvent(data []byte) *WindowFocusEvent {
	s := string(data)
	switch s {
	case focusInSeq:
		return &WindowFocusEvent{Focused: true}
	case focusOutSeq:
		return &WindowFocusEvent{Focused: false}
	}
	return nil
}
