package lnwire

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
)

type MessageDirection int

const (
	MessageSent     MessageDirection = 1
	MessageReceived MessageDirection = 2
)

func (d MessageDirection) String() string {
	switch d {
	case MessageSent:
		return "sent"
	case MessageReceived:
		return "received"
	default:
		return "<error, unknown direction>"
	}
}

type MessageInfo struct {
	Msg        Message
	PeerPubKey *btcec.PublicKey
	Direction  MessageDirection // "received" or "sent"
	Time       time.Time
}

func (mi *MessageInfo) MarshalJSON() ([]byte, error) {
	type MessageInfoJSON struct {
		MsgRaw     string `json:"msg_raw"`
		PeerPubKey string `json:"peer_pubkey"`
		Direction  string `json:"direction"`
		Type       string `json:"type"`
		Time       string `json:"time"`
	}
	b := new(bytes.Buffer)
	if _, err := WriteMessage(b, mi.Msg, 0); err != nil {
		return nil, fmt.Errorf("cannot encode message: %v", err)
	}
	pubKeyStr := ""
	if mi.PeerPubKey != nil {
		pubKeyStr = hex.EncodeToString(mi.PeerPubKey.SerializeCompressed())
	}
	msg := MessageInfoJSON{
		MsgRaw:     hex.EncodeToString(b.Bytes()),
		PeerPubKey: pubKeyStr,
		Direction:  mi.Direction.String(),
		Type:       mi.Msg.MsgType().String(),
		Time:       fmt.Sprint(mi.Time.Unix()),
	}
	return json.Marshal(&msg)
}
