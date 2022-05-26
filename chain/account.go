package chain

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"time"
)

type Account struct {
	ID        []byte
	CreatedAt time.Time
	PublicKey *btcec.PublicKey
}

func CreateAccount(privateKey *btcec.PrivateKey, specialID []byte) (account *Account) {
	_, AccountPublicKey := btcec.PrivKeyFromBytes(privateKey.Serialize())
	accountId := AccountPublicKey.SerializeCompressed()
	if len(specialID) > 0 {
		copy(accountId, specialID)
	}
	account = &Account{
		ID:        accountId,
		PublicKey: AccountPublicKey,
		CreatedAt: time.Now(),
	}
	return account
}
