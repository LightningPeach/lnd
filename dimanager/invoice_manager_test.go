package dimanager

import (
	"crypto/sha256"
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"
	"testing"
	"time"
	"bytes"
)

// getHandler returns handler for a given list of preimages
func getHandler(rPreImages [][32]byte) *Handler {
	// Plan:
	// 1. Create list of rhash
	// 2. Create map of rhash -> Invoice
	// 3.Launch goroutine
	// 4. return channels

	rHashToInvoice := make(map[[32]byte]*lnrpc.Invoice)
	for index, rPreImg := range rPreImages {
		rHash := sha256.Sum256(rPreImg[:])
		// Because we pass slice we need to make local copy rPreImg
		// So slice will be different in each iteration
		rPreImg := rPreImg
		invoice := &lnrpc.Invoice{
			RHash:     rHash[:],
			RPreimage: rPreImg[:],
			Memo:      fmt.Sprintf("invoice %v", index),
		}
		rHashToInvoice[rHash] = invoice
	}

	chRHash := make(chan [32]byte)
	chInvRez := make(chan *InvoiceResult)

	go func() {
		defer close(chInvRez)
		for rHash := range chRHash {
			if inv, ok := rHashToInvoice[rHash]; ok {
				chInvRez <- &InvoiceResult{
					RHash:   rHash,
					Invoice: inv,
				}
			} else {
				chInvRez <- &InvoiceResult{
					RHash:   rHash,
					Invoice: nil,
				}
			}
		}
	}()

	return &Handler{
		ChRHash:   chRHash,
		ChInvoice: chInvRez,
	}
}

func TestDynamicInvoiceManager_RegisterHandler(t *testing.T) {
	preImgs := [][32]byte{
		UintTo32Byte(1),
		UintTo32Byte(2),
		UintTo32Byte(3),
	}
	dim := NewDynamicInvoiceManager()
	err := dim.Start()
	if err != nil {
		t.Fatalf("Cannot start invoice manager: %v", err)
	}

	inv, err := dim.GetInvoice(UintTo32Byte(0), time.Second)
	if inv != nil || err == nil {
		t.Fatalf("incorect result without handlers")
	}

	handler1 := getHandler(preImgs)
	handler1Id, err := dim.RegisterHandler(handler1)
	if err != nil {
		t.Fatalf("cannot register handler")
	}

	for ind, rPreImg := range preImgs {
		rHash := sha256.Sum256(rPreImg[:])
		inv, err := dim.GetInvoice(rHash, time.Second)
		if inv == nil || err != nil {
			t.Fatalf("incorect getinvoice for existing invoice")
		}
		if inv.Memo != fmt.Sprintf("invoice %v", ind) || !bytes.Equal(rPreImg[:], inv.RPreimage) || !bytes.Equal(rHash[:], inv.RHash) {
			t.Fatalf("incorrect invoice")
		}
	}

	inv, err = dim.GetInvoice(UintTo32Byte(4), time.Second)
	if inv != nil || err == nil {
		t.Fatalf("incorect result without handlers")
	}
	if inv != nil || err == nil {
		t.Fatalf("incorect result for not existing invoice")
	}

	// Now add one more handler
	preImgs2 := [][32]byte{
		UintTo32Byte(4),
		UintTo32Byte(5),
		UintTo32Byte(6),
	}

	handler2 := getHandler(preImgs2)
	handler2Id, err := dim.RegisterHandler(handler2)
	if err != nil {
		t.Fatalf("cannot register second handler")
	}

	if handler1Id == handler2Id {
		t.Fatalf("each handler shoudl have uniqu id")
	}

	for i:=0; i<6; i++ {
		var ind int
		var rPreImg [32]byte

		if i < 3 {
			ind = i
			rPreImg = preImgs[ind]
		} else {
			ind = i - 3
			rPreImg = preImgs2[ind]
		}
		rHash := sha256.Sum256(rPreImg[:])
		inv, err := dim.GetInvoice(rHash, time.Second)
		if inv == nil || err != nil {
			t.Fatalf("incorect getinvoice for existing invoice")
		}
		if inv.Memo != fmt.Sprintf("invoice %v", ind) || !bytes.Equal(rPreImg[:], inv.RPreimage) || !bytes.Equal(rHash[:], inv.RHash) {
			t.Fatalf("incorrect invoice")
		}
	}
}

func TestDynamicInvoiceManager_UnregisterHandler(t *testing.T) {
	preImgs := [][32]byte{
		UintTo32Byte(1),
		UintTo32Byte(2),
		UintTo32Byte(3),
	}
	dim := NewDynamicInvoiceManager()
	err := dim.Start()
	if err != nil {
		t.Fatalf("Cannot start invoice manager: %v", err)
	}

	inv, err := dim.GetInvoice(UintTo32Byte(0), time.Second)
	if inv != nil || err == nil {
		t.Fatalf("incorect result without handlers")
	}

	handler1 := getHandler(preImgs)
	handler1Id, err := dim.RegisterHandler(handler1)
	if err != nil {
		t.Fatalf("cannot register handler")
	}

	for ind, rPreImg := range preImgs {
		rHash := sha256.Sum256(rPreImg[:])
		inv, err := dim.GetInvoice(rHash, time.Second)
		if inv == nil || err != nil {
			t.Fatalf("incorect getinvoice for existing invoice")
		}
		if inv.Memo != fmt.Sprintf("invoice %v", ind) || !bytes.Equal(rPreImg[:], inv.RPreimage) || !bytes.Equal(rHash[:], inv.RHash) {
			t.Fatalf("incorrect invoice")
		}
	}

	err = dim.UnregisterHandler(handler1Id)
	if err != nil {
		t.Fatalf("cannot unregister handler: %v", err)
	}

	for _, rPreImg := range preImgs {
		rHash := sha256.Sum256(rPreImg[:])
		inv, err := dim.GetInvoice(rHash, time.Second)
		if inv != nil || err == nil {
			t.Fatalf("should not find invoice")
		}
	}
}

func TestDynamicInvoiceManager_RegisterHandler2(t *testing.T) {
	// Test if some handler hangs up
	preImgs := [][32]byte{
		UintTo32Byte(1),
		UintTo32Byte(2),
		UintTo32Byte(3),
	}
	dim := NewDynamicInvoiceManager()
	err := dim.Start()
	if err != nil {
		t.Fatalf("Cannot start invoice manager: %v", err)
	}

	inv, err := dim.GetInvoice(UintTo32Byte(0), time.Second)
	if inv != nil || err == nil {
		t.Fatalf("incorect result without handlers")
	}

	handler1 := getHandler(preImgs)
	_, err = dim.RegisterHandler(handler1)
	if err != nil {
		t.Fatalf("cannot register handler")
	}

	// This handler receives nothing so any send to it will block
	handler2 := &Handler{
		ChInvoice: make(chan *InvoiceResult),
		ChRHash: make(chan [32]byte),
	}
	_, err = dim.RegisterHandler(handler2)
	if err != nil {
		t.Fatalf("cannot register handler")
	}

	for ind, rPreImg := range preImgs {
		rHash := sha256.Sum256(rPreImg[:])
		inv, err := dim.GetInvoice(rHash, time.Second)
		if inv == nil || err != nil {
			t.Fatalf("incorect getinvoice for existing invoice")
		}
		if inv.Memo != fmt.Sprintf("invoice %v", ind) || !bytes.Equal(rPreImg[:], inv.RPreimage) || !bytes.Equal(rHash[:], inv.RHash) {
			t.Fatalf("incorrect invoice")
		}
	}
}