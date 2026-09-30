package builder

import "errors"

type errWriter struct {
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

type shortWriter struct {
	limit int
	buf   []byte
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p)
	if n > w.limit {
		n = w.limit
	}
	w.buf = append(w.buf, p[:n]...)
	return n, nil
}

type zeroWriter struct{}

func (w *zeroWriter) Write(p []byte) (int, error) {
	return 0, nil
}

type failAfterWriter struct {
	remaining int
	err       error
	buf       []byte
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, w.err
	}
	n := len(p)
	if n > w.remaining {
		n = w.remaining
	}
	w.buf = append(w.buf, p[:n]...)
	w.remaining -= n
	if w.remaining == 0 {
		return n, w.err
	}
	return n, nil
}

var errTest = errors.New("test write error")
