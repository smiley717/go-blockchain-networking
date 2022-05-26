package network

import (
	"blockchain/chain"
	"encoding/binary"
)

// Commands between peers
var (
	// Peer List Management
	CommandAnnouncement uint8 = 0 // Announcement
	CommandResponse     uint8 = 1 // Response
	CommandPing         uint8 = 2 // Keep-alive message (no payload).
	CommandPong         uint8 = 3 // Response to ping (no payload).
	// Blockchain
	CommandGetBlock uint8 = 4 // Request blocks for specified peer.

)

type AnnouncementPayload struct {
	Features          uint8  // 0:1 Feature support
	Port              uint16 // 1:3 External port if known. 0 if not.
	BlockchainVersion uint64 // 3:7 Blockchain version
	BlockchainHeight  uint64 // 7:11 Blockchain height
	PortInternal      uint16 // 11:13 Internal port. Can be used to detect NATs.
}

func EncodeAnnouncement(node *chain.Node, sequence uint32) (packetBody *PacketBody) {
	packetBody = new(PacketBody)
	payload := make([]byte, 20)
	payload[0] = node.FeaturesSupport()
	binary.BigEndian.PutUint16(payload[1:1+2], node.Port)
	binary.BigEndian.PutUint64(payload[3:3+8], node.BlockchainVersion)
	binary.BigEndian.PutUint64(payload[11:11+8], node.BlockchainHeight)
	packetBody.Command = CommandAnnouncement
	packetBody.Protocol = 0
	packetBody.Payload = payload
	packetBody.Sequence = sequence
	return packetBody
}
