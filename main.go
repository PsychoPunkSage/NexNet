package main

import (
	"fmt"
	"log"

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

	s1.Start()
}

func makeServer(listenAddr string, nodes ...string) *server.FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: Onpeer func
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := server.FileServerOpts{
		StorageRoot:       listenAddr[1:] + "_network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	return server.NewFileServer(fileServerOpts)
}
