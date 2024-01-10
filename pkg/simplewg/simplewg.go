package simplewg

import "sync"

type Wg sync.WaitGroup

func (wg *Wg) Go(fn func()) {
	wg.Wg().Add(1)
	go func() {
		defer wg.Wg().Done()
		fn()
	}()
}

func (wg *Wg) Wg() *sync.WaitGroup { return (*sync.WaitGroup)(wg) }
func (wg *Wg) Wait()               { wg.Wg().Wait() }
