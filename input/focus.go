package input

// WindowFocusEvent — окно терминала получило или потеряло фокус
// (DECSET 1004, CSI I / CSI O).
type WindowFocusEvent struct {
	Focused bool
}

const (
	focusInSeq  = "\x1b[I"
	focusOutSeq = "\x1b[O"
)

// ParseFocus разбирает последовательность фокуса окна.
func ParseFocus(data []byte) *WindowFocusEvent {
	switch string(data) {
	case focusInSeq:
		return &WindowFocusEvent{Focused: true}
	case focusOutSeq:
		return &WindowFocusEvent{Focused: false}
	}
	return nil
}
