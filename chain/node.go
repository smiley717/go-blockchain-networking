package chain

import (
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
)

// Features are sent as bit array in the Announcement message.
const (
	FeatureValidator = 0 // Sender is a validator
	FeatureIndexer   = 1 // Sender is an indexer
)

type Node struct {
	ID                []byte
	PublicKey         *btcec.PublicKey
	Port              uint16
	IsValidator       bool
	IsIndexer         bool
	BlockchainHeight  uint64 // Blockchain height
	BlockchainVersion uint64 // Blockchain version
}

func (node *Node) FeaturesSupport() (features byte) {
	if node.IsValidator {
		features |= 1 << FeatureValidator
	}
	if node.IsIndexer {
		features |= 1 << FeatureIndexer
	}
	return features
}

func (node *Node) String() string {
	return fmt.Sprintf("ID= %X, Port= %d, IsValidator= %t, IsIndexer= %t, BlockchainHeight= %d, BlockchainVersion=%d",
		node.ID, node.Port, node.IsValidator, node.IsIndexer, node.BlockchainHeight, node.BlockchainVersion,
	)
}
