package dimanager

import (
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"
)

// pendingRequests represents information about pending invoice requests,
// to what handlers they were sent and from what handlers response obtained
type pendingRequests struct {
	invoiceRequests              []*invoiceRequestCmd
	sentToHandlers               map[HandlerId]struct{}
	receivedResponseFromHandlers map[HandlerId]struct{}
}

func newPendingRequests() *pendingRequests {
	return &pendingRequests{
		invoiceRequests:              make([]*invoiceRequestCmd, 0),
		sentToHandlers:               make(map[HandlerId]struct{}),
		receivedResponseFromHandlers: make(map[HandlerId]struct{}),
	}
}

func (pr *pendingRequests) addInvoiceRequest(req *invoiceRequestCmd) {
	pr.invoiceRequests = append(pr.invoiceRequests, req)
}

func (pr *pendingRequests) sendResult(inv *lnrpc.Invoice) {
	for _, invReqCmd := range pr.invoiceRequests {
		invReqCmd.chRez <- inv
		invReqCmd.chErr <- nil
		close(invReqCmd.chDone)
	}
}

func (pr *pendingRequests) sendError(err error) {
	for _, invReqCmd := range pr.invoiceRequests {
		invReqCmd.chRez <- nil
		invReqCmd.chErr <- err
		close(invReqCmd.chDone)
	}
}

func (pr *pendingRequests) sentToHandler(handlerId HandlerId) {
	pr.sentToHandlers[handlerId] = struct{}{}
}

func (pr *pendingRequests) receivedResponseFromHandler(handlerId HandlerId) {
	pr.receivedResponseFromHandlers[handlerId] = struct{}{}
}

func (pr *pendingRequests) isReceivedFromAll(existingHandlers map[HandlerId]*handlerExtended) bool {
	for id := range pr.sentToHandlers {
		_, exist := existingHandlers[id]
		_, receivedResponse := pr.receivedResponseFromHandlers[id]
		if exist && !receivedResponse {
			return false
		}
	}
	return true
}

func (pr *pendingRequests) isEmpty() bool {
	return len(pr.invoiceRequests) == 0
}

func (pr *pendingRequests) handleTimeout(msg *invoiceRequestCmd) {
	for i, invReqCmd := range pr.invoiceRequests {
		if invReqCmd == msg {
			msg.chErr <- fmt.Errorf("timeout while waiting for result")
			msg.chRez <- nil
			close(msg.chDone)
			// Delete element i from pr.invoiceRequests
			pr.invoiceRequests = append(pr.invoiceRequests[:i], pr.invoiceRequests[i+1:]...)
		}
	}
}