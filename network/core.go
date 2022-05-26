package network

import (
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/panjf2000/gnet/v2"
	"log"
)

var (
	server TcpServer
)

func BootStrap(privateKey *btcec.PrivateKey, publicKey *btcec.PublicKey) {
	var port int
	var multicore bool

	flag.IntVar(&port, "port", 9000, "--port 9000")
	flag.BoolVar(&multicore, "multicore", true, "--multicore true")
	flag.Parse()
	server = TcpServer{
		multicore:   multicore,
		port:        uint16(port),
		PrivateKey:  privateKey,
		PublicKey:   publicKey,
		LookupTable: new(LookupTable),
	}
	err := gnet.Run(&server, fmt.Sprintf("tcp://:%d", port), gnet.WithMulticore(multicore), gnet.WithTicker(true))
	if err != nil {
		log.Printf("server exits with error: %v", err)
		panic(err.Error())
	}
}
