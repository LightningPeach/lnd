package dimanager

import (
	"fmt"
	"time"
	"github.com/lightningnetwork/lnd/lnrpc"
	"sync/atomic"
)

type invoiceRequestCmd struct {
	rhash   [32]byte
	timeout time.Duration
	chRez   chan *lnrpc.Invoice
	chErr   chan error
	// chDone is closed when processing request is done
	// so goroutines can exit
	chDone chan struct{}
}

type registerHandlerCmd struct {
	handler *Handler
	chRez   chan HandlerId
	chErr   chan error
}

type unregisterHandlerCmd struct {
	handlerId HandlerId
	chErr     chan error
}

type invoiceResultExtended struct {
	InvoiceResult
	handlerId HandlerId
}

type handlerExtended struct {
	Handler
	chStop chan struct{}
	rHashInfChannel InfiniteBuffer
}

type dynamicInvoiceManager struct {
	chCmd           chan interface{}
	chStop          chan struct{}
	chTimeout       chan *invoiceRequestCmd
	chInvoiceResult chan *invoiceResultExtended
	isStopped int32

	handlers               map[HandlerId]*handlerExtended
	nextHandlerId          HandlerId
	pendingInvoiceRequests map[[32]byte]*pendingRequests
}

func (di *dynamicInvoiceManager) newTimeout(req *invoiceRequestCmd, chExit <-chan struct{}) {
	go func() {
		select {
		case <-time.After(req.timeout):
			di.chTimeout <- req
		case <-chExit:
		case <-di.chStop:
		}
	}()
}

func (di *dynamicInvoiceManager) newReceiver(chInvoiceResults chan *InvoiceResult, handlerId HandlerId, chExit chan struct{}) {
	go func() {
		for {
			select {
			// TODO(mkl): should i check if channel is actually closed
			case inv := <-chInvoiceResults:
				if inv == nil {
					fmt.Println("warn: got nil in invoice result")
					continue
				}
				if !inv.isValid() {
					fmt.Println("invalid invoice")
					continue
				}
				invExt := &invoiceResultExtended{
					*inv,
					handlerId,
				}
				di.chInvoiceResult <- invExt
			case <-di.chStop:
				return
			case <-chExit:
				return
			}
		}
	}()
}

func (di *dynamicInvoiceManager) handleInvoiceRequestCmd(msg *invoiceRequestCmd) {
	if len(di.handlers) == 0 {
		msg.chRez <- nil
		msg.chErr <- fmt.Errorf("no handlers available")
		close(msg.chDone)
		return
	}
	pending, ok := di.pendingInvoiceRequests[msg.rhash]
	if ok {
		// Some requests for this rhash was already sent,
		// so simply add it to waiting list
		pending.addInvoiceRequest(msg)
		di.newTimeout(msg, msg.chDone)
	} else {
		// No waiting requests with given rhash exist so need to send it
		pending = newPendingRequests()
		pending.addInvoiceRequest(msg)
		di.pendingInvoiceRequests[msg.rhash] = pending
		// TODO(mkl): correctly handle timeout
		di.newTimeout(msg, msg.chDone)
		// TODO(mkl): may be do it unblocking
		for handlerId, handler := range di.handlers {
			handler.rHashInfChannel.ChIn() <- msg.rhash
			pending.sentToHandler(handlerId)
		}
	}
}

func (di *dynamicInvoiceManager) handleRegisterHandlerCmd(msg *registerHandlerCmd) {
	handlerExt := handlerExtended{
		Handler: *msg.handler,
		chStop:  make(chan struct{}),
		rHashInfChannel: NewInfiniteBuffer(msg.handler.ChRHash),
	}
	id := di.nextHandlerId
	di.nextHandlerId += 1
	di.handlers[id] = &handlerExt
	di.newReceiver(msg.handler.ChInvoice, id, handlerExt.chStop)
	msg.chRez <- id
	msg.chErr <- nil
}

func (di *dynamicInvoiceManager) handleUnregisterHandlerCmd(msg *unregisterHandlerCmd) {
	handlerExt, ok := di.handlers[msg.handlerId]
	if ok {
		close(handlerExt.chStop)
		handlerExt.rHashInfChannel.ForceClose()
		delete(di.handlers, msg.handlerId)
		msg.chErr <- nil
	} else {
		msg.chErr <- fmt.Errorf("handler with given id=%v do not exist", msg.handlerId)
	}
}

func (di *dynamicInvoiceManager) handleInvoiceResult(invRez *invoiceResultExtended) {
	pendingRequests, ok := di.pendingInvoiceRequests[invRez.RHash]
	if !ok {
		return
	}
	if invRez.Invoice != nil {
		pendingRequests.sendResult(invRez.Invoice)
		delete(di.pendingInvoiceRequests, invRez.RHash)
	} else {
		pendingRequests.receivedResponseFromHandler(invRez.handlerId)
		if pendingRequests.isReceivedFromAll(di.handlers) {
			pendingRequests.sendError(fmt.Errorf("all handlers do not have the invoice"))
			delete(di.pendingInvoiceRequests, invRez.RHash)
		}
	}
}

func (di *dynamicInvoiceManager) handleTimeout(invReqCmd *invoiceRequestCmd) {
	pending, ok := di.pendingInvoiceRequests[invReqCmd.rhash]
	if !ok {
		return
	}
	pending.handleTimeout(invReqCmd)
	if pending.isEmpty() {
		delete(di.pendingInvoiceRequests, invReqCmd.rhash)
	}
}

func (di *dynamicInvoiceManager) handleStop() {
	for _, pending := range di.pendingInvoiceRequests {
		pending.sendError(fmt.Errorf("dynamic invoice manager was stoped"))
	}
}

func (di *dynamicInvoiceManager) mainLoop() {
	for {
		select {
		case cmd := <-di.chCmd:
			switch msg := cmd.(type) {
			case *invoiceRequestCmd:
				di.handleInvoiceRequestCmd(msg)
			case *registerHandlerCmd:
				di.handleRegisterHandlerCmd(msg)
			case *unregisterHandlerCmd:
				di.handleUnregisterHandlerCmd(msg)
			default:
				// TODO(mkl): add error handling or should it panic?
				fmt.Printf("internal error: incorrect message type: %T\n", cmd)
			}
		case invRez := <-di.chInvoiceResult:
			di.handleInvoiceResult(invRez)
		case invReqCmd := <-di.chTimeout:
			di.handleTimeout(invReqCmd)
		case <-di.chStop:
			di.handleStop()
			break
		}
	}
}

func (di *dynamicInvoiceManager) Start() error {
	go di.mainLoop()
	return nil
}

func (di *dynamicInvoiceManager) Stop() error {
	// TODO(mlk): should i wait until all spawned goroutines exit
	if atomic.CompareAndSwapInt32(&di.isStopped, 1, 0) {
		close(di.chStop)
	}
	return nil
}

func (di *dynamicInvoiceManager) GetInvoice(rHash [32]byte, timeout time.Duration) (*lnrpc.Invoice, error) {
	req := &invoiceRequestCmd{
		rhash:   rHash,
		timeout: timeout,
		chRez:   make(chan *lnrpc.Invoice, 1),
		chErr:   make(chan error, 1),
		chDone:  make(chan struct{}),
	}
	di.chCmd <- req
	return <-req.chRez, <-req.chErr
}

func (di *dynamicInvoiceManager) RegisterHandler(handler *Handler) (HandlerId, error) {
	if handler == nil {
		return 0, fmt.Errorf("handler should not be nil")
	}
	if handler.ChInvoice == nil || handler.ChRHash == nil {
		return 0, fmt.Errorf("channels in handler should not be nil")
	}
	req := &registerHandlerCmd{
		handler: handler,
		chRez:   make(chan HandlerId, 1),
		chErr:   make(chan error, 1),
	}
	di.chCmd <- req
	return <-req.chRez, <-req.chErr
}

func (di *dynamicInvoiceManager) UnregisterHandler(handlerId HandlerId) error {
	req := &unregisterHandlerCmd{
		handlerId: handlerId,
		chErr:     make(chan error, 1),
	}
	di.chCmd <- req
	return <-req.chErr
}
