package parallel

import (
	"sync"
	"sync/atomic"
)

type Group struct {
	Queue     *Queue
	lastError atomic.Value
	wait      sync.WaitGroup
	lock      sync.Mutex
}

func (q *Queue) NewGroup() *Group {
	return &Group{Queue: q}
}

func (g *Group) Add(f func() error) {
	g.wait.Add(1)
	g.Queue.Add(func() error {
		err := f()
		if err != nil {
			g.lastError.Store(err)
		}

		g.wait.Done()

		return err
	})
}

// Wait for all the tasks to finish
func (g *Group) Wait() bool {
	g.wait.Wait()
	return g.lastError.Load() == nil
}

// Syncronized uses a per-group mutex to lock access to some client-provided data, presumably a map or slice of results.
func (g *Group) Sync(f func()) {
	g.lock.Lock()
	defer g.lock.Unlock()
	f()
}
