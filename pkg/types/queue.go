package types

import (
	"fmt"
	"sync"
)

// FIFO with unlimited capacity and internal buffer access
// not thread safe
type queue struct {
	data []interface{}
}

func (q *queue) len() int {
	return len(q.data)
}

func (q *queue) push(v interface{}) {
	q.data = append(q.data, v)
}

// panics if empty
func (q *queue) pop() interface{} {
	v := q.data[0]
	q.data = q.data[1:]
	return v
}

func (q *queue) getData() []interface{} {
	return q.data
}

// 1 ctrl M send N recv
type ControlledQueue struct {
	data          queue
	mu            sync.Mutex
	requestRecvCh chan struct{}
	stopCh        chan struct{}
}

func NewControlledQueue() *ControlledQueue {
	return &ControlledQueue{
		stopCh:        make(chan struct{}),
		requestRecvCh: make(chan struct{}, 1),
	}
}

// closes underlying channel
// only call once from ctrl
func (cq *ControlledQueue) Close() {
	cq.mu.Lock()
	close(cq.stopCh)
	close(cq.requestRecvCh)
	cq.mu.Unlock()
}

// return true on Send
// return false if closed and not Send
func (cq *ControlledQueue) Send(v interface{}) bool {
	select {
	case <-cq.stopCh:
		return false
	default:
		//fmt.Println("cq send, waiting to lock...")
		cq.mu.Lock()
		//fmt.Println("cq send, len:", cq.dataQ.len(), "->", cq.dataQ.len()+1)
		if cq.data.len() == 0 || len(cq.requestRecvCh) < 1 {
			select {
			case cq.requestRecvCh <- struct{}{}:
			default:
			}
		}
		cq.data.push(v)
		//fmt.Println("cq", cq.dataQ.len())
		cq.mu.Unlock()
		return true
	}
}

// blocks on empty to wait to receive
func (cq *ControlledQueue) Recv() (interface{}, bool) {
	_, v, ok := cq.AttemptRecv(true)
	return v, ok
}

// return (false, nil, true) on empty
// return (true, v, true) on recv
// return (true, nil, false) on closed
// can opt out of blocking on empty
func (cq *ControlledQueue) AttemptRecv(blockOnEmpty bool) (canRecv bool, v interface{}, ok bool) {
	for {
		select {
		case <-cq.stopCh:
			fmt.Println("cq attempt recv, closed")
			return true, nil, false
		default:
		}

		//fmt.Println("cq attempt recv, waiting to lock...")
		cq.mu.Lock()
		//fmt.Println("cq attempt recv, len:", cq.dataQ.len())
		if cq.data.len() > 0 {
			break
		}
		if blockOnEmpty == false {
			cq.mu.Unlock()
			//fmt.Println("cq attempt recv failed")
			return false, nil, true
		}
		cq.mu.Unlock()
		//fmt.Println("wait for value...")
		<-cq.requestRecvCh
		//fmt.Println("got value")
	}

	//fmt.Println("cq recv")
	v = cq.data.pop()
	//fmt.Println("cq", cq.dataQ.len())
	cq.mu.Unlock()
	return true, v, true
}

func (cq *ControlledQueue) Lock() []interface{} {
	cq.mu.Lock()
	return cq.data.data
}

func (cq *ControlledQueue) Unlock() {
	cq.mu.Unlock()
}
