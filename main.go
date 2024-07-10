package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/PsychoPunkSage/NexNet/cryptography"
	"github.com/PsychoPunkSage/NexNet/p2p"
	"github.com/PsychoPunkSage/NexNet/server"
	"github.com/PsychoPunkSage/NexNet/storage"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	fmt.Println("Doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {
	s := makeServer(":3000", "")
	s1 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)

	go s1.Start()
	time.Sleep(1 * time.Second)

	for i := 0; i < 10; i++ {
		data := bytes.NewReader([]byte(fmt.Sprintf("A very big data file %d", i)))
		s1.Store(fmt.Sprintf("PrivateData%d", i), data)
		time.Sleep(5 * time.Millisecond)
	}

	key := "PrivateData1"
	if err := s1.Remove(key); err != nil {
		log.Fatal(err)
	}

	r, err := s1.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received: %s\n", b)

	select {}
}

func makeServer(listenAddr string, nodes ...string) *server.FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := server.FileServerOpts{
		EncKey:            cryptography.NewEncryptionKey(),
		StorageRoot:       listenAddr[1:] + "_network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := server.NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}
