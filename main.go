package main

import (
	"blockchain/chain"
	"blockchain/config"
	"blockchain/network"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"log"
	"os"
	"time"
)

func main() {
	nodeConfig := new(config.Config)
	if status, err := config.LoadConfig("nodeConfig.yml", nodeConfig); status != config.ExitSuccess {
		log.Printf("Enable to load nodeConfig file: status = %d, error = %v", status, err)
	}
	if len(nodeConfig.PrivateKey) == 0 {
		log.Printf("Init: must provide Private Key \n")
		os.Exit(config.ExitPrivateKeyCreate)
	}
	// load existing key from Config, if available
	configPK, err := hex.DecodeString(nodeConfig.PrivateKey)
	if err != nil {
		log.Printf("Init: private key in Config is corrupted! Error: %s\n", err.Error())
		os.Exit(config.ExitPrivateKeyCorrupt)
	}

	PrivateKey, PublicKey := btcec.PrivKeyFromBytes(configPK)
	// BlockChain
	_, err = chain.BootStrap()
	if err != nil {
		log.Printf("main -> error: %s", err.Error())
		os.Exit(config.ExitBlockchainCorrupt)
	}

	// Network
	network.BootStrap(PrivateKey, PublicKey)

	for {
		time.Sleep(1e8)
	}
}
