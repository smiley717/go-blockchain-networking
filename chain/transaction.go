package chain

import (
	"blockchain/hash"
	"bytes"
	"encoding/gob"
	"log"
)

const (
	TransactionStatusUnknown uint8 = 0
)

type Transaction struct {
	ID        []byte
	Type      uint16
	Status    uint8
	Signature []byte
	Timestamp uint64
}

func (transaction *Transaction) Hash() []byte {
	txCopy := *transaction
	txCopy.ID = []byte{}
	return hash.HashData(txCopy.Serialize())
}

func (transaction Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(transaction)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func Deserialize(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}
	return transaction
}
