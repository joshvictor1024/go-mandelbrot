package main

import (
	"fmt"

	"github.com/joshvictor1024/go-mandelbrot/pkg/types"
)

// push chunk to dirty if not in dirty
type iterateQueue struct {
	cq    *types.ControlledQueue
	dirty map[*chunk]struct{}
}

func newIterateQueue(dirty map[*chunk]struct{}) *iterateQueue {
	return &iterateQueue{
		cq:    types.NewControlledQueue(),
		dirty: dirty,
	}
}

func (iq *iterateQueue) close() {
	iq.cq.Close()
}

// if chunk exist in data (dirty)
// replace old iw with new
// return false to signal close (like with channels)
func (iq *iterateQueue) send(iw *iterateWork) bool {
	if _, ok := iq.dirty[iw.chunk]; ok {
		data := iq.cq.Lock()
		for i, v := range data {
			oldIw := v.(*iterateWork)
			if oldIw.chunk == iw.chunk {
				data[i] = iw
				fmt.Println("dirty")
				iq.cq.Unlock()
				return true
			}
		}
		iq.cq.Unlock()
		fmt.Println("dirty but not in iq")
		return true
	}
	iq.dirty[iw.chunk] = struct{}{}
	return iq.cq.Send(iw)
}

func (iq *iterateQueue) recv() (*iterateWork, bool) {
	v, ok := iq.cq.Recv()
	if ok {
		return v.(*iterateWork), ok
	}
	return nil, ok
}

// 1 ctrl M send 1 recv
type drawQueue struct {
	cq *types.ControlledQueue
}

func newDrawQueue() *drawQueue {
	return &drawQueue{
		cq: types.NewControlledQueue(),
	}
}

// call from control
// close underlying channel
// only call once
func (dq *drawQueue) close() {
	dq.cq.Close()
}

// return true on send
// return false if closed and not send
func (dq *drawQueue) send(dw *drawWork) bool {
	return dq.cq.Send(dw)
}

func (dq *drawQueue) recv() (*drawWork, bool) {
	v, ok := dq.cq.Recv()
	if ok {
		return v.(*drawWork), ok
	}
	return nil, ok
}

func (dq *drawQueue) attemptRecv(blockOnEmpty bool) (bool, *drawWork, bool) {
	canRecv, v, ok := dq.cq.AttemptRecv(blockOnEmpty)
	if canRecv && ok {
		return canRecv, v.(*drawWork), ok
	}
	return canRecv, nil, ok
}

// 1 ctrl M send 1 recv
type iterationBufferQueue struct {
	cq *types.ControlledQueue
}

func newIterationBufferQueue() *iterationBufferQueue {
	return &iterationBufferQueue{
		cq: types.NewControlledQueue(),
	}
}

// call from control
// close underlying channel
// only call once
func (ibq *iterationBufferQueue) close() {
	ibq.cq.Close()
}

func (ibq *iterationBufferQueue) send(ib *iterationBuffer) bool {
	return ibq.cq.Send(ib)
}

func (ibq *iterationBufferQueue) recv() (*iterationBuffer, bool) {
	ib, ok := ibq.cq.Recv()
	if ok {
		return ib.(*iterationBuffer), ok
	}
	return nil, ok
}
