package parallel

import (
	"sync"
)

type Queue struct {
	capacity   int
	running    int
	pending    []func() error
	firstError error
	mutex      sync.Mutex
	wait       sync.WaitGroup
}

func New(capacity int) *Queue {
	return &Queue{
		capacity: capacity,
	}
}

func (q *Queue) Add(f func() error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.wait.Add(1)

	if q.running < q.capacity {
		q.running++
		go q.execute(f)
	} else {
		q.pending = append(q.pending, f)
	}
}

func (q *Queue) Wait() error {
	q.wait.Wait()
	return q.firstError
}

func (q *Queue) execute(f func() error) {
	err := f()

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if err != nil && q.firstError == nil {
		q.firstError = err
	}

	if len(q.pending) == 0 {
		q.running--
		q.pending = nil // release the memory used by the slice's underlying error
	} else {
		f := q.pending[0]
		q.pending = q.pending[1:]
		go q.execute(f)
	}

	q.wait.Done()
}