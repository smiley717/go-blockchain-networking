package network

import (
	"github.com/panjf2000/gnet/v2"
	"time"
)

type Peer struct {
	gnet.Conn
	ConnectionTime time.Time
	LastSeen       time.Time
	Authenticated  bool
}

func (peer *Peer) ShouldMaintain() bool {
	if !peer.Authenticated && time.Now().Sub(peer.ConnectionTime).Seconds() > 10 {
		err := peer.Close()
		if err != nil {
			println("Error closing peer Connection", peer.String())
		}
		return false
	}
	return true
}
func (peer *Peer) String() string {
	return peer.RemoteAddr().String()
}
