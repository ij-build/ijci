package util

import "io"

type chanReader struct {
	ch       <-chan []byte
	halt     chan (struct{})
	overflow []byte
}

func NewChanReader(ch <-chan []byte) io.ReadCloser {
	return &chanReader{
		ch:   ch,
		halt: make(chan struct{}),
	}
}

func (r *chanReader) Read(p []byte) (int, error) {
	if len(r.overflow) > 0 {
		return r.sendChunk(p, r.overflow)
	}

	select {
	case chunk := <-r.ch:
		return r.sendChunk(p, chunk)
	case <-r.halt:
	}

	return 0, io.EOF
}

func (r *chanReader) sendChunk(p, chunk []byte) (int, error) {
	if len(chunk) == 0 {
		return 0, io.EOF
	}

	if len(chunk) > len(p) {
		r.overflow = append(r.overflow, chunk[len(p):]...)
		chunk = chunk[:len(p)]
	}

	copy(p, chunk)
	return len(chunk), nil
}

func (r *chanReader) Close() error {
	close(r.halt)
	return nil
}
