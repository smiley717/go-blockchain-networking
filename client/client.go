package main

import (
	"blockchain/chain"
	"blockchain/network"
	"bufio"
	"encoding/hex"
	"flag"
	"github.com/btcsuite/btcd/btcec/v2"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2/pkg/logging"
)

func logErr(err error) {
	logging.Error(err)
	if err != nil {
		log.Printf("Error %v", err.Error())
	}
}

func main() {
	var (
		network     string
		addr        string
		concurrency int
		packetSize  int
		packetBatch int
		packetCount int
	)

	// Example command: go run client.go --network tcp --address ":9000" --concurrency 100 --packet_size 1024 --packet_batch 20 --packet_count 1000
	flag.StringVar(&network, "network", "tcp", "--network tcp")
	flag.StringVar(&addr, "address", "127.0.0.1:9000", "--address 127.0.0.1:9000")
	flag.IntVar(&concurrency, "concurrency", 1024, "--concurrency 500")
	flag.IntVar(&packetSize, "packet_size", 1024, "--packe_size 256")
	flag.IntVar(&packetBatch, "packet_batch", 100, "--packe_batch 100")
	flag.IntVar(&packetCount, "packet_count", 10000, "--packe_count 10000")
	flag.Parse()

	logging.Infof("start %d clients...", concurrency)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			runClient(network, addr, packetSize, packetBatch, packetCount)
			wg.Done()
		}()
	}
	wg.Wait()
	logging.Infof("all %d clients are done", concurrency)
}

func runClient(network, addr string, packetSize, batch, count int) {
	rand.Seed(time.Now().UnixNano())
	c, err := net.Dial(network, addr)
	logErr(err)
	logging.Infof("connection=%s starts...", c.LocalAddr().String())
	defer func() {
		logging.Infof("connection=%s stops...", c.LocalAddr().String())
		c.Close()
	}()
	rd := bufio.NewReader(c)

	for i := 0; i < count; i++ {
		batchSendAndRecv(c, rd, packetSize, batch)
	}
	for {
		// maintain connection
		time.Sleep(1e8)
	}
}

func batchSendAndRecv(c net.Conn, rd *bufio.Reader, packetSize, batch int) {
	codec := network.Codec{}

	var (
		requests  [][]byte
		buf       []byte
		packetLen int
	)
	for i := 0; i < batch; i++ {
		req := make([]byte, packetSize)
		_, err := rand.Read(req)
		logErr(err)
		requests = append(requests, req)
		/*packet := protocol.PacketBody{
			Protocol: 1,
			Command:  0x00,
			Sequence: 0x00,
			Payload:  req,
		}*/
		receiverPrivateKey, _ := hex.DecodeString("1E99423A4ED27608A15A2616A2B0E9E52CED330AC530EDCC32C8FFC6A526AEDD")
		_, ReceiverPublicKey := btcec.PrivKeyFromBytes(receiverPrivateKey)

		senderPrivateKey, _ := hex.DecodeString("E9873D79C6D87DC0FB6A5778633389F4453213303DA61F20BD67FC233AA33262")
		SenderPrivateKey, SenderPublicKey := btcec.PrivKeyFromBytes(senderPrivateKey)

		node := chain.Node{
			Port:              123,
			IsIndexer:         true,
			IsValidator:       false,
			PublicKey:         SenderPublicKey,
			BlockchainVersion: 1,
			BlockchainHeight:  233,
		}
		data, _ := codec.Encode(SenderPrivateKey, ReceiverPublicKey, network.EncodeAnnouncement(&node, 123))
		packetLen = len(data)
		buf = append(buf, data...)
	}
	log.Printf("writing %X", buf)
	_, err := c.Write(buf)
	logErr(err)
	respPacket := make([]byte, batch*packetLen)
	for len(respPacket) == 0 {
		_, err = io.ReadFull(rd, respPacket)
		if err == nil {
			for i, _ := range requests {
				rsp, err := codec.Unpack(respPacket[i*packetLen:])
				logErr(err)
				log.Printf("received response %X", rsp)
			}
		}
		if err != io.EOF {
			logErr(err)
		}
		time.Sleep(1e8)
	}

}
