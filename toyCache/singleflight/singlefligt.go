package singleflight

import "sync"

// call is an in-flight(calling) or completed Do call
type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

// Group represent a class of work and forms a namespace in which
// unit of work can be executed with duplicate suppression
type Group struct {
	mu 	sync.Mutex			// protects m
	m 	map[string]*call	// lazily initialization
}

// Do executes and return the result of given function, making sure that
// only one executed in flight for a given key at a time.
// If a duplicate comes in, the duplicate caller waits for the previous caller
// complete and receives the same results
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}