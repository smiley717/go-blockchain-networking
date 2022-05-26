package network

import (
	"blockchain/chain"
	"encoding/binary"
	"log"
)

func ProcessPacket(packet *IncomingPacket) {
	codec := packet.Peer.Context().(*Codec)
	packetBody := packet.Body
	switch packet.Body.Command {
	case CommandAnnouncement:
		announcement := AnnouncementPayload{
			Features:          packetBody.Payload[0],
			Port:              binary.BigEndian.Uint16(packetBody.Payload[1:3]),
			BlockchainVersion: binary.BigEndian.Uint64(packetBody.Payload[3 : 3+8]),
			BlockchainHeight:  binary.BigEndian.Uint64(packetBody.Payload[11 : 11+8]),
		}
		log.Printf("[%X]: ProcessPacket -> Announcement", packet.NodeID)
		node := chain.Node{
			ID:                packet.NodeID,
			PublicKey:         packet.PublicKey,
			BlockchainHeight:  announcement.BlockchainHeight,
			BlockchainVersion: announcement.BlockchainVersion,
			Port:              announcement.Port,
			IsValidator:       announcement.Features&(1<<chain.FeatureValidator) > 0,
			IsIndexer:         announcement.Features&(1<<chain.FeatureIndexer) > 0,
		}
		log.Printf("[%X]: ProcessPacket -> Announcement from %s", packet.NodeID, node.String())
		announcementResponse := EncodeAnnouncement(server.Node, 2)
		response, err := codec.Encode(server.PrivateKey, packet.PublicKey, announcementResponse)
		if err != nil {
			log.Printf("[%X]: ProcessPacket -> Error to answer %s", packet.NodeID, node.String())
			return
		}
		packet.Peer.Write(response)
	}
}
