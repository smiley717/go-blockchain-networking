package hash

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"lukechampine.com/blake3"
)

// HashData returns blake3 hash which is 32bytes = 256bits hash.
func HashData(data []byte) (hash []byte) {
	hash32 := blake3.Sum256(data)
	return hash32[:]
}

// PublicKey2NodeID translates the Public Key into the node ID.
func PublicKey2NodeID(publicKey *btcec.PublicKey) (nodeID []byte) {
	return HashData(publicKey.SerializeCompressed())
}
