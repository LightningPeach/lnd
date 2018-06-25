package dimanager

import (
	"sync"
	"testing"
	"time"
)

func TestInfiniteBuffer_Len(t *testing.T) {
	ch := make(chan [32]byte)
	chInf := NewInfiniteBuffer(ch)
	l := chInf.Len()
	if l != 0 {
		t.Errorf("Initial len should be 0, got %v", l)
	}

	chInf.ChIn() <- UintTo32Byte(1)
	time.Sleep(10 * time.Millisecond)
	l = chInf.Len()
	if l != 1 {
		t.Errorf("After pushing len should be 1, got %v", l)
	}

	chInf.ChIn() <- UintTo32Byte(2)
	time.Sleep(10 * time.Millisecond)
	l = chInf.Len()
	if l != 2 {
		t.Errorf("After pushing len should be 1, got %v", l)
	}

	val := <-ch
	if val != UintTo32Byte(1) {
		t.Errorf("Output channel produce wrong value")
	}
	time.Sleep(10 * time.Millisecond)
	l = chInf.Len()
	if l != 1 {
		t.Errorf("Len should be 1, got %v", l)
	}

	val = <-ch
	if val != UintTo32Byte(2) {
		t.Errorf("Output channel produce wrong value")
	}
	time.Sleep(10 * time.Millisecond)
	l = chInf.Len()
	if l != 0 {
		t.Errorf("Len should be 1, got %v", l)
	}
}

func TestInfiniteBuffer_Sequential(t *testing.T) {
	ch := make(chan [32]byte)
	chInf := NewInfiniteBuffer(ch)
	for i := 0; i < 100; i++ {
		chInf.ChIn() <- UintTo32Byte(uint(i))
	}
	for i := 0; i < 100; i++ {
		val := <-ch
		if val != UintTo32Byte(uint(i)) {
			t.Errorf("Incorrect value from the channel")
		}
	}
}

func TestInfiniteBuffer_Parallel(t *testing.T) {
	ch := make(chan [32]byte)
	chInf := NewInfiniteBuffer(ch)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			chInf.ChIn() <- UintTo32Byte(uint(i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			val := <-ch
			if val != UintTo32Byte(uint(i)) {
				t.Errorf("Incorrect value from the channel")
			}
		}
	}()
	wg.Wait()
}

func TestInfiniteBuffer_ChInClose(t *testing.T) {
	ch := make(chan [32]byte)
	chInf := NewInfiniteBuffer(ch)

	for i := 0; i < 5; i++ {
		chInf.ChIn() <- UintTo32Byte(uint(i))
	}
	close(chInf.ChIn())
	for i := 0; i < 5; i++ {
		val := <-ch
		if val != UintTo32Byte(uint(i)) {
			t.Errorf("Incorrect value from the channel")
		}
	}
	_, ok := <- ch
	if ok {
		t.Errorf("output channel should be closed")
	}
	if chInf.Len() != 0 {
		t.Errorf("length of closed channel should be 0")
	}
}

func TestInfiniteBuffer_ForceClose(t *testing.T) {
	ch := make(chan [32]byte)
	chInf := NewInfiniteBuffer(ch)

	for i := 0; i < 5; i++ {
		chInf.ChIn() <- UintTo32Byte(uint(i))
	}
	chInf.ForceClose()
	time.Sleep(10 * time.Millisecond)
	_, ok := <- ch
	if ok {
		t.Errorf("output channel should be closed")
	}
	if chInf.Len() != 0 {
		t.Errorf("length of closed channel should be 0")
	}
}
