package network

import (
	"blockchain/chain"
	"blockchain/hash"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/panjf2000/gnet/v2"
	"log"
	"time"
)

type TcpServer struct {
	gnet.BuiltinEventEngine

	engine    gnet.Engine
	port      uint16
	multicore bool

	Node        *chain.Node
	PrivateKey  *btcec.PrivateKey
	PublicKey   *btcec.PublicKey
	LookupTable *LookupTable
}

func (server *TcpServer) OnBoot(engine gnet.Engine) gnet.Action {
	server.engine = engine
	server.Node = new(chain.Node)
	server.Node.PublicKey = server.PublicKey
	server.Node.ID = hash.PublicKey2NodeID(server.Node.PublicKey)
	server.Node.Port = server.port
	log.Printf("Server Node public key: %X", server.Node.PublicKey.SerializeCompressed())
	log.Printf("Server Node ID: %X", server.Node.ID)
	log.Printf("TCP server with multi-core=%t is listening on %s\n", server.multicore, fmt.Sprintf("tcp://:%d", server.port))
	return gnet.None
}

func (server *TcpServer) OnOpen(connection gnet.Conn) (out []byte, action gnet.Action) {
	connection.SetContext(new(Codec))

	log.Printf("OnOpen: connected peers %d", server.engine.CountConnections())
	peer := &Peer{Conn: connection, ConnectionTime: time.Now()}
	server.LookupTable.add(peer)
	return
}

func (server *TcpServer) OnClose(connection gnet.Conn, err error) (action gnet.Action) {
	log.Printf("[%s]: OnClose -> connection closed", connection.RemoteAddr().String())
	if err != nil {
		log.Printf("[%s]: OnClose -> error occurred on connection, %v\n", connection.RemoteAddr().String(), err)
	}
	peer := server.LookupTable.peers[connection.RemoteAddr().String()]
	if peer != nil {
		server.LookupTable.remove(peer)
	}
	log.Printf("OnClose: connected peers %d", server.engine.CountConnections())

	return gnet.None
}

func (server *TcpServer) OnTraffic(connection gnet.Conn) gnet.Action {
	log.Printf("[%s]: OnTraffic -> buffered %d bytes", connection.RemoteAddr().String(), connection.InboundBuffered())

	codec := connection.Context().(*Codec)
	peer := server.LookupTable.peers[connection.RemoteAddr().String()]
	if peer == nil {
		log.Printf("[%s]: OnTraffic -> peer not found closing connection", connection.RemoteAddr().String())
		return gnet.Close
	}
	packet, err := codec.Decode(peer, server.Node.PublicKey)
	if err == ErrorIncompletePacket {
		return gnet.None
	}
	if err != nil {
		log.Printf("[%s]: OnTraffic -> invalid packet %v", connection.RemoteAddr().String(), err)
		return gnet.Close
	}
	log.Printf("[%s]: OnTraffic ->  %x", connection.RemoteAddr().String(), packet.Body.String())
	if packet.Body.Command == CommandAnnouncement {
		log.Printf("[%s]: OnTraffic -> is Authenticated", connection.RemoteAddr().String())
		peer.Authenticated = true
	}
	go ProcessPacket(packet)
	return gnet.None
}

func (server *TcpServer) OnTick() (delay time.Duration, action gnet.Action) {
	for _, peer := range server.LookupTable.peers {
		maintain := peer.ShouldMaintain()
		if !maintain {
			n, err := peer.Write([]byte("Timeout"))
			if err != nil || n != 7 {
				log.Printf("Error sending timeout message to peer %s: %v", peer.String(), err)
			}
			log.Printf("Closing connection to %s for authentication timeout", peer.String())
			err = peer.Close()
			if err != nil {
				log.Printf("Error closing connection with peer %s after timeout", peer.String())
			}
		}
	}
	return time.Second, gnet.None
}
