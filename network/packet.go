package network

import (
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"time"
)

// PacketBody is a decrypted P2P message
type PacketBody struct {
	Protocol uint8  // Protocol version = 0
	Command  uint8  // Command code
	Sequence uint32 // Sequence number
	Payload  []byte // Payload
}

type OutboundPacket struct {
	Body      PacketBody
	PublicKey *btcec.PublicKey
}

type IncomingPacket struct {
	Peer       *Peer
	Body       PacketBody
	PublicKey  *btcec.PublicKey
	NodeID     []byte
	ReceivedAt time.Time
}

func (packetBody *PacketBody) String() string {
	return fmt.Sprintf("Protcol: %d, Command: %d, Sequence: %d, Payload: %x", packetBody.Protocol, packetBody.Command, packetBody.Sequence, packetBody.Payload)
}
