package dimanager

import (
	"bytes"
	"crypto/sha256"
	"github.com/lightningnetwork/lnd/lnrpc"
	"time"
	"fmt"
)

// InvoiceResult return result of finding invoice
// if Invoice is nil than no invoices has been found for a given hash
type InvoiceResult struct {
	// RHash is 32-byte hash(R) for invoice
	RHash [32]byte
	// Invoice is Lighning invoice
	Invoice *lnrpc.Invoice
}

func (ir *InvoiceResult) isValid() bool {
	if ir.Invoice == nil {
		return true
	}
	if !bytes.Equal(ir.Invoice.RHash, ir.RHash[:]) {
		fmt.Println("Hash not equal")
		return false
	}
	hashPreimg := sha256.Sum256(ir.Invoice.RPreimage)
	if !bytes.Equal(ir.RHash[:], hashPreimg[:]) {
		fmt.Println("incorrect preimage")
		return false
	}
	return true
}

// Handler represents connection to external service for
// generating invoices.
// External service should read hashes from ChRHash and send found/not found
// invoices to ChInvoice
type Handler struct {
	ChRHash   chan [32]byte
	ChInvoice chan *InvoiceResult
}

type HandlerId int64

type DynamicInvoiceManager interface {
	Start() error
	GetInvoice(rHash [32]byte, timeout time.Duration) (*lnrpc.Invoice, error)
	//  ChRHash may be closed during operation to indicate that no input is obtained
	RegisterHandler(handler *Handler) (HandlerId, error)
	UnregisterHandler(handlerId HandlerId) error
	Stop() error
}

func NewDynamicInvoiceManager() DynamicInvoiceManager {
	return &dynamicInvoiceManager{
		chCmd:                  make(chan interface{}),
		chStop:                 make(chan struct{}),
		chTimeout:              make(chan *invoiceRequestCmd),
		chInvoiceResult:        make(chan *invoiceResultExtended),
		isStopped:              0,
		handlers:               make(map[HandlerId]*handlerExtended),
		nextHandlerId:          0,
		pendingInvoiceRequests: make(map[[32]byte]*pendingRequests),
	}
}
