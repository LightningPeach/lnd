package dimanager

import "sync/atomic"

// InfiniteBuffer represents channel with infinite buffer size
// It is prefered not to use such approach. In some instances it may be
// useful e.g. when sending messages to some external system and it should
// never block. It contains some methods for dealing with such situations
// Example: you have channel chA and you want that operation chA <- x
// never blocks, so use create infinite buffer
// chABuff := NewInfinityBuffer(chA)
// and use chABuff.ChIn() instead of chA for sending values.
// 1. If chABuff.ChIn() is closed (behavior is similar to normal buffered channels)
//   1.1 All values from buffer will be send to chA. Then chA will be closed
//   1.2 If there are no values in buffer, chA will be closed
// 2. Len() - returns number of elements in a buffer. However chABuff.ChIn() <- val
//    do not immediately change Len(). Some time is required because updating done in
//    loop in separate goroutine.
// 3. ForceClose() - closes output channel and discards buffer
type InfiniteBuffer interface {
	ChIn() chan <- [32]byte
	Len() int
	ForceClose()
}

// NewInfiniteBuffer creates new infinite buffer from given output channel
func NewInfiniteBuffer(chOut chan [32]byte) InfiniteBuffer {
	ib := &infiniteBuffer{
		chIn:       make(chan [32]byte),
		chOut:      chOut,
		chClose:    make(chan struct{}),
		buffer:     make([][32]byte, 0),
		isInClosed: false,
		isStopped: 0,
	}
	go ib.mainLoop()
	return ib
}

type infiniteBuffer struct {
	chIn       chan [32]byte
	chOut      chan [32]byte
	chClose    chan struct{}
	buffer     [][32]byte
	isInClosed bool
	isStopped int32
}

func (ib *infiniteBuffer) ChIn() chan<- [32]byte {
	return ib.chIn
}

func (ib *infiniteBuffer) Len() int {
	return len(ib.buffer)
}

func (ib *infiniteBuffer) ForceClose() {
	if atomic.CompareAndSwapInt32(&ib.isStopped, 0, 1) {
		close(ib.chClose)
	}
}

func (ib *infiniteBuffer) mainLoop() {
	defer func() {
		atomic.StoreInt32(&ib.isStopped, 1)
	}()
	for {
		// There are 4 situations depending on 2 criteria
		// 1. Is input channel closed? If it is closed we cannot
		//    select on it because it always returns.
		// 2. Is something in a buffer? If there are nothing in a buffer we cannot
		//    select on output channel because we have no value to send
		if ib.isInClosed {
			// We cannot receive new values because
			// input channel is closed. We can only
			// send values from buffer.
			if len(ib.buffer) == 0 {
				close(ib.chOut)
				return
			} else {
				select {
				case ib.chOut <- ib.buffer[0]:
					ib.buffer = ib.buffer[1:]
				case <-ib.chClose:
					close(ib.chOut)
					ib.buffer = nil
					return
				}
			}
		} else {
			// Input channel is not closed, so we can receive values
			if len(ib.buffer) == 0 {
				// We can only receive
				select {
				case x, ok := <-ib.chIn:
					if !ok {
						ib.isInClosed = true
					} else {
						ib.buffer = append(ib.buffer, x)
					}
				case <-ib.chClose:
					close(ib.chOut)
					ib.buffer = nil
					return
				}
			} else {
				// We can both send and receive
				select {
				case x, ok := <-ib.chIn:
					if !ok {
						ib.isInClosed = true
					} else {
						ib.buffer = append(ib.buffer, x)
					}
				case ib.chOut <- ib.buffer[0]:
					ib.buffer = ib.buffer[1:]
				case <-ib.chClose:
					close(ib.chOut)
					ib.buffer = nil
					return
				}
			}
		}
	}
}

func UintTo32Byte(a uint) [32]byte {
	var rez [32]byte
	i := 0
	for a > 0 {
		rez[i] = byte(a % 256)
		i += 1
		a /= 256
	}
	return rez
}