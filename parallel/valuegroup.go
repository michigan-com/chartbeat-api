package parallel

import (
	"sync"
)

type ValueGroup struct {
	Group

	nextOrd      int32
	results      map[interface{}]interface{}
	slice        []interface{}
	resultsMutex sync.Mutex
}

func (q *Queue) NewValueGroup() *ValueGroup {
	return &ValueGroup{Group: Group{Queue: q}}
}

func (g *ValueGroup) Add(key interface{}, f func() (interface{}, error)) {
	g.resultsMutex.Lock()
	defer g.resultsMutex.Unlock()

	idx := len(g.slice)
	g.slice = append(g.slice, nil)

	g.Group.Add(func() error {
		result, err := f()

		g.resultsMutex.Lock()
		defer g.resultsMutex.Unlock()

		if key != nil {
			if g.results == nil {
				g.results = make(map[interface{}]interface{})
			}
			g.results[key] = result
		}

		g.slice[idx] = result

		return err
	})
}

func (g *ValueGroup) WaitMap() (map[interface{}]interface{}, error) {
	g.Group.Wait()

	g.resultsMutex.Lock()
	defer g.resultsMutex.Unlock()

	return g.results, g.lastError.Load().(error)
}

func (g *ValueGroup) WaitSlice() ([]interface{}, error) {
	g.Group.Wait()

	g.resultsMutex.Lock()
	defer g.resultsMutex.Unlock()

	return g.slice, g.lastError.Load().(error)
}
