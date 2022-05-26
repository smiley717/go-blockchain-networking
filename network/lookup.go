package network

import (
	"log"
	"sync"
)

type LookupTable struct {
	peers     map[string]*Peer
	listMutex sync.RWMutex
}

func (lut *LookupTable) size() uint16 {
	return uint16(len(lut.peers))
}

func (lut *LookupTable) add(peer *Peer) {
	lut.listMutex.Lock()
	defer lut.listMutex.Unlock()
	if lut.peers == nil {
		lut.peers = make(map[string]*Peer)
	}
	lut.peers[peer.String()] = peer
	log.Printf("peer added: %s", peer.String())
}

func (lut *LookupTable) remove(peer *Peer) {
	lut.listMutex.Lock()
	defer lut.listMutex.Unlock()
	if _, ok := lut.peers[peer.String()]; ok {
		delete(lut.peers, peer.String())
		log.Printf("peer removed: %s", peer.String())
	}
}
