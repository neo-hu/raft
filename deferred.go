package raft

import (
	"context"
	"sync"
)

type deferredResult struct {
	err    error
	result interface{}
}

type deferred struct {
	req      interface{}
	resultCh chan struct{}
	result   *deferredResult
	sync.RWMutex
	once sync.Once
	ctx  context.Context
}

func newDeferred(ctx context.Context, req interface{}) *deferred {
	return &deferred{
		req:      req,
		ctx:      ctx,
		resultCh: make(chan struct{}, 1),
	}
}

func (d *deferred) ResultCh() <-chan struct{} {
	return d.resultCh
}

func (d *deferred) Result() (interface{}, error) {
	d.RLock()
	defer d.RUnlock()
	if d.result == nil {
		return nil, nil
	}
	return d.result.result, d.result.err
}

func (d *deferred) done(result interface{}, err error) {
	d.once.Do(func() {
		d.Lock()
		defer d.Unlock()
		d.result = &deferredResult{
			result: result,
			err:    err,
		}
		d.resultCh <- struct{}{}
		close(d.resultCh)
	})

}
