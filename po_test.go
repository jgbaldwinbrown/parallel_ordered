package po

import (
	"log"
	"testing"
	"sync"
)

func TestOrderer(t *testing.T) {
	o := NewOrderer[IndexedVal[float64]](0)
	var wg sync.WaitGroup
	arr := make([]float64, 0, 2000)

	wg.Add(2000)
	for i := 0; i < 2000; i++ {
		i := i
		go func() {
			o.Write(IndexedVal[float64]{I: i, Val: float64(i)})
			wg.Done()
		}()
	}
	done := make(chan struct{})
	go func() {
		for val, ok := o.Read(); ok; val, ok = o.Read() {
			arr = append(arr, val.Val)
		}
		done <- struct{}{}
	}()

	log.Println("starting wait")
	wg.Wait()
	log.Println("finished wait")
	o.Close()
	log.Println("finished close")
	<-done
	log.Println("finished done")

	if len(arr) != 2000 {
		t.Errorf("len(arr) %v != 2000", len(arr))
	}
	for i, f := range arr {
		if i != int(f) {
			t.Errorf("i %v != int(f) %v", i, int(f))
		}
	}
}
