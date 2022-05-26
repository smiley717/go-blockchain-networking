package chain

import (
	"blockchain/store"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"
)

// StatusX provides information about the blockchain status. Some errors code indicate a corruption.
const (
	StatusOK            = 0 // No problems in the blockchain detected.
	StatusBlockNotFound = 1 // Missing block in the blockchain.
	StatusCorruptBlock  = 2 // Error block encoding
)

const (
	HeightOffset = 0
	HeightSize   = 8

	VersionOffset = HeightOffset + HeightSize
	VersionSize   = 8

	FormatOffset = VersionOffset + VersionSize
	FormatSize   = 2

	HeaderSize = FormatOffset + FormatSize
)

// Blockchain stores the blockchain's header in memory. Any changes must be synced to disk!
type Blockchain struct {
	// header
	height  uint64 // [0:8] Height is exchanged as uint32 in the protocol, but stored as uint64.
	version uint64 // [8:16] Version is always uint64.
	format  uint16 // [16:18] Format is only locally used.

	accounts map[string]*Account
	// internals
	path       string      // Path of the blockchain on disk. Depends on key-value store whether a filename or folder.
	database   store.Store // The database storing the blockchain.
	sync.Mutex             // synchronized access to the header

	// callback
	BlockchainUpdate func(blockchain *Blockchain, oldHeight, oldVersion, newHeight, newVersion uint64)
}

// BootStrap initializes the blockchain. It creates the blockchain database file if it does not exist already.
func BootStrap() (blockchain *Blockchain, err error) {
	var dbPath = "/tmp/blockchain/db"
	blockchain = &Blockchain{path: dbPath}

	// open existing blockchain file or create new one
	if blockchain.database, err = store.NewPogrebStore(dbPath); err != nil {
		return nil, err
	}

	// verify header
	var found bool

	found, err = blockchain.headerRead()
	if err != nil {
		return blockchain, err // likely corrupt blockchain database
	} else if !found {
		if err := blockchain.headerWrite(0, 0); err != nil {
			return blockchain, err
		}
	}

	log.Printf("Blockchain -> bootstraped height=%d, version=%d", blockchain.height, blockchain.version)
	return blockchain, nil
}

// the key names in the key-value database are constant and must not collide with block numbers (i.e. they must be >64 bit)
const keyHeader = "header"

// headerRead reads the header from the blockchain and decodes it.
func (blockchain *Blockchain) headerRead() (found bool, err error) {
	buffer, found := blockchain.database.Get([]byte(keyHeader))
	if !found {
		return false, nil
	}

	if len(buffer) != HeaderSize {
		return true, errors.New("blockchain header size mismatch")
	}

	blockchain.height = binary.BigEndian.Uint64(buffer[HeightOffset:VersionOffset])
	blockchain.version = binary.BigEndian.Uint64(buffer[VersionOffset:FormatOffset])

	return
}

// headerWrite writes the header to the blockchain and signs it.
func (blockchain *Blockchain) headerWrite(height, version uint64) (err error) {
	oldHeight := blockchain.height
	oldVersion := blockchain.version

	blockchain.height = height
	blockchain.version = version

	var buffer [HeaderSize]byte
	binary.BigEndian.PutUint64(buffer[HeightOffset:VersionOffset], height)
	binary.BigEndian.PutUint64(buffer[VersionOffset:FormatOffset], version)
	binary.BigEndian.PutUint16(buffer[FormatOffset:HeaderSize], 0) // Current format is 0

	err = blockchain.database.Set([]byte(keyHeader), buffer[:])

	// call the callback, if any
	if blockchain.BlockchainUpdate != nil {
		blockchain.BlockchainUpdate(blockchain, oldHeight, oldVersion, blockchain.height, blockchain.version)
	}

	return err
}

func (blockchain *Blockchain) AddAccount(account *Account) {
	if prevAccount := blockchain.accounts[fmt.Sprintf("%x.web3", account.ID)]; prevAccount != nil {

	}
}
