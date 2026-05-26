package audio

import (
	"io"
	"sync"
)

type PCMQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	buf    []byte
	closed bool
	limit  int
}

func NewPCMQueue(limit int) *PCMQueue {
	q := &PCMQueue{limit: limit}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *PCMQueue) Write(p []byte) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return 0, io.ErrClosedPipe
	}
	q.buf = append(q.buf, p...)
	if q.limit > 0 && len(q.buf) > q.limit {
		excess := len(q.buf) - q.limit
		q.buf = append([]byte(nil), q.buf[excess:]...)
	}
	q.cond.Broadcast()
	return len(p), nil
}

func (q *PCMQueue) Read(p []byte) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.buf) == 0 && !q.closed {
		q.cond.Wait()
	}
	if len(q.buf) == 0 && q.closed {
		return 0, io.EOF
	}
	n := copy(p, q.buf)
	copy(q.buf, q.buf[n:])
	q.buf = q.buf[:len(q.buf)-n]
	return n, nil
}

func (q *PCMQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.closed {
		q.closed = true
		q.cond.Broadcast()
	}
	return nil
}
