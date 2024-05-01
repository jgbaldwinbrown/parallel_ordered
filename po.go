package po

import (
	"fmt"
	"sync"
)

type Indexer interface {
	Index() int
}

type IndexedVal[T any] struct {
	Val T
	I int
}

func (i IndexedVal[T]) Index() int {
	return i.I
}

type Orderer[T Indexer] struct {
	M map[int]T
	Pos int
	Closed bool
	Mu sync.Mutex
	Cond *sync.Cond
	ReadMu sync.Mutex
	WriteMu sync.Mutex
}

func NewOrderer[T Indexer](cap int) *Orderer[T] {
	o := new(Orderer[T])
	o.M = make(map[int]T, cap)
	o.Cond = sync.NewCond(&o.Mu)
	return o
}

func (o *Orderer[T]) Write(val T) {
	o.WriteMu.Lock()
	defer o.WriteMu.Unlock()

	o.Mu.Lock()
	defer o.Mu.Unlock()

	i := val.Index()
	if i < o.Pos {
		panic(fmt.Errorf("Orderer.Write: i %v < o.Pos %v", i, o.Pos))
	}
	o.M[i] = val

	if i == o.Pos {
		o.Cond.Broadcast()
	}
}

func (o *Orderer[T]) Read() (T, bool) {
	o.ReadMu.Lock()
	defer o.ReadMu.Unlock()

	o.Mu.Lock()
	defer o.Mu.Unlock()

	for {
		_, ok := o.M[o.Pos]
		if o.Closed || ok {
			break
		}
		o.Cond.Wait()
	}

	if o.Closed && len(o.M) < 1 {
		var t T
		return t, false
	}

	for {
		val, ok := o.M[o.Pos]
		if ok {
			delete(o.M, o.Pos)
			o.Pos++
			return val, true
		}
		o.Pos++
	}
	var t T
	return t, false
}

func (o *Orderer[T]) Close() {
	o.Mu.Lock()
	defer o.Mu.Unlock()
	o.Closed = true
	o.Cond.Broadcast()
}
