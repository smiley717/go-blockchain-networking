package network

import (
	"blockchain/hash"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"golang.org/x/crypto/salsa20"
	"log"
	"math/rand"
	"time"
)

var ErrorIncompletePacket = errors.New("INCOMPLETE PACKET")

/*
Offset  Size   Info
0		2	   Magic Number
2       4      Nonce
6       1      Protocol version = 0
7       1      Command
8       4      Sequence
12      2      Size of payload data
14      ?      Payload
        ?      Randomized garbage
?		65     Signature, ECDSA secp256k1 512-bit + 1 header byte
*/
const (
	magicNumberSize     = 2
	nonceSize           = 4
	protocolVersionSize = 1
	commandSize         = 1
	sequenceSize        = 4
	payloadLengthSize   = 2
	signatureSize       = 65

	magicNumberOffset     = 0
	nonceOffset           = magicNumberOffset + magicNumberSize
	protocolVersionOffset = nonceOffset + nonceSize
	commandOffset         = protocolVersionOffset + protocolVersionSize
	sequenceOffset        = commandOffset + commandSize
	payloadLengthOffset   = sequenceOffset + sequenceSize
	payloadOffset         = payloadLengthOffset + payloadLengthSize

	magicNumber   = 0x2424
	maxBodyLength = 1030
)
const PacketLengthMin = payloadOffset + signatureSize
const maxRandomGarbage = 20

var magicNumberBytes []byte

func init() {
	magicNumberBytes = make([]byte, magicNumberSize)
	binary.BigEndian.PutUint16(magicNumberBytes, uint16(magicNumber))
}

type Codec struct {
}

// Encode encrypts a packet using the provided senders private key and receivers compressed public key.
func (codec Codec) Encode(senderPrivateKey *btcec.PrivateKey, receiverPublicKey *btcec.PublicKey, packet *PacketBody) ([]byte, error) {
	garbage := packetGarbage(maxRandomGarbage)
	log.Printf("Encode -> garbage: %x", garbage)

	data := make([]byte, PacketLengthMin+len(packet.Payload)+len(garbage))

	// add magic number and nonce to header
	binary.BigEndian.PutUint16(data[magicNumberOffset:nonceOffset], magicNumber)
	log.Printf("Encode -> magic number: %x", data[magicNumberOffset:nonceOffset])

	nonce := rand.Uint32()
	nonceB := make([]byte, 8)
	binary.BigEndian.PutUint32(nonceB[4:8], nonce)

	copy(data[nonceOffset:protocolVersionOffset], nonceB[4:8])
	log.Printf("Encode -> nonce: %x", data[nonceOffset:protocolVersionOffset])

	// populate body
	data[protocolVersionOffset] = packet.Protocol
	data[commandOffset] = packet.Command

	binary.BigEndian.PutUint32(data[sequenceOffset:payloadOffset], packet.Sequence)
	binary.BigEndian.PutUint16(data[payloadLengthOffset:payloadOffset], uint16(len(packet.Payload)))

	copy(data[payloadOffset:], packet.Payload)

	garbageOffset := payloadOffset + len(packet.Payload)
	copy(data[garbageOffset:garbageOffset+len(garbage)], garbage)

	log.Printf("Encode -> body: %x", data[protocolVersionOffset:garbageOffset+len(garbage)])

	log.Printf("Encode -> data: %x", data[:len(data)-signatureSize])

	// encrypt body using Salsa20
	keySalsa := publicKeyToSalsa20Key(receiverPublicKey)
	log.Printf("Encode -> key salsa: %x", keySalsa)
	salsa20.XORKeyStream(data[protocolVersionOffset:garbageOffset+len(garbage)], data[protocolVersionOffset:garbageOffset+len(garbage)], nonceB, keySalsa)

	log.Printf("Encode -> data encrypted: %x", data[:len(data)-signatureSize])

	signature, e := ecdsa.SignCompact(senderPrivateKey, hash.HashData(data[:len(data)-signatureSize]), true)

	log.Printf("Encode -> signature: %x", signature)

	if e != nil {
		return nil, e
	}
	// encrypt signature using Salsa20
	salsa20.XORKeyStream(signature[:], signature[:], nonceB, keySalsa)

	log.Printf("Encode -> signature encrypted: %x", signature)

	copy(data[len(data)-signatureSize:], signature)

	return data, nil
}

func (codec *Codec) Decode(peer *Peer, receiverPublicKey *btcec.PublicKey) (packet *IncomingPacket, err error) {
	receivedAt := time.Now()
	raw, _ := peer.Peek(-1)

	log.Printf("[%s]: Decode -> raw= %X", peer.RemoteAddr().String(), raw)

	if peer.InboundBuffered() < PacketLengthMin {
		log.Printf("[%s]: Decode -> ErrorIncompletePacket buffered %d minimum expected %d", peer.RemoteAddr().String(), peer.InboundBuffered(), PacketLengthMin)
		return nil, ErrorIncompletePacket
	}
	if !bytes.Equal(magicNumberBytes, raw[magicNumberOffset:nonceOffset]) {
		err = errors.New(fmt.Sprintf("INVALID MAGIC NUMBER: Expected '%s' but got '%s'", magicNumberBytes, raw[magicNumberOffset:nonceOffset]))
		return nil, err
	}

	nonce := make([]byte, nonceSize+4)
	copy(nonce[4:8], raw[nonceOffset:protocolVersionOffset])

	// Verify the signature and extract the public key from it.
	var signature [signatureSize]byte
	copy(signature[:], raw[len(raw)-signatureSize:])
	keySalsa := publicKeyToSalsa20Key(receiverPublicKey)
	salsa20.XORKeyStream(signature[:], signature[:], nonce, keySalsa)

	senderPublicKey, _, err := ecdsa.RecoverCompact(signature[:], hash.HashData(raw[:len(raw)-signatureSize]))
	if err != nil {
		return nil, err
	}
	log.Printf("[%s]: Decode -> SenderPublicKey= %X", peer.RemoteAddr().String(), senderPublicKey.SerializeCompressed())

	// Decrypt the packet using Salsa20.
	bufferBodyDecrypted := make([]byte, len(raw)-protocolVersionOffset-signatureSize) // full length -signature -nonce - magic number
	salsa20.XORKeyStream(bufferBodyDecrypted[:], raw[protocolVersionOffset:len(raw)-signatureSize], nonce, keySalsa)
	log.Printf("[%s]: Decode -> Decrypted IncomingPacket= %x", peer.RemoteAddr().String(), bufferBodyDecrypted)

	packetBody := PacketBody{Protocol: bufferBodyDecrypted[0], Command: bufferBodyDecrypted[commandOffset-protocolVersionOffset]}
	packetBody.Sequence = binary.BigEndian.Uint32(bufferBodyDecrypted[sequenceOffset-protocolVersionOffset : payloadLengthOffset-protocolVersionOffset])

	payloadLength := binary.BigEndian.Uint16(bufferBodyDecrypted[payloadLengthOffset-protocolVersionOffset : payloadOffset-protocolVersionOffset])

	if payloadLength > maxBodyLength {
		log.Printf("[%s]: Decode -> msgLength %d > max allowed %d", peer.RemoteAddr().String(), payloadLength, maxBodyLength)
		peer.Discard(peer.InboundBuffered())
	}

	if payloadLength > 0 {
		packetBody.Payload = make([]byte, payloadLength)
		copy(packetBody.Payload, bufferBodyDecrypted[payloadOffset-protocolVersionOffset:payloadOffset-protocolVersionOffset+payloadLength])
	}

	peer.Discard(len(raw))

	packet = &IncomingPacket{Peer: peer, Body: packetBody, PublicKey: senderPublicKey, NodeID: hash.PublicKey2NodeID(senderPublicKey), ReceivedAt: receivedAt}
	log.Printf("[%s]: Decode -> Received IncomingPacket Body= %s", peer.RemoteAddr().String(), packet.Body.String())
	return packet, nil
}

func (codec Codec) Unpack(buffer []byte) ([]byte, error) {
	if len(buffer) < PacketLengthMin {
		return nil, ErrorIncompletePacket
	}
	return buffer, nil
}

func packetGarbage(packetLength int) (random []byte) {
	b := make([]byte, rand.Intn(packetLength))
	if _, err := rand.Read(b); err != nil {
		return nil
	}
	return b
}

func publicKeyToSalsa20Key(publicKey *btcec.PublicKey) (key *[32]byte) {
	// bit 0 from PublicKey.Y is ignored here, but is negligible for this purpose
	key = new([32]byte)
	copy(key[:], publicKey.SerializeCompressed()[1:])
	return key
}
