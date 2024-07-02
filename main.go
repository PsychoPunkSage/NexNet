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
	// tcpOpts := p2p.TCPTransportOpts{
	// 	ListenAddr:    ":3000",
	// 	HandshakeFunc: p2p.NOPHandshakeFunc,
	// 	Decoder:       p2p.DefaultDecoder{},
	// 	OnPeer:        OnPeer,
	// }
	// tr := p2p.NewTCPTransport(tcpOpts)

	// fmt.Println("AP is here..")

	// go func() {
	// 	for {
	// 		msg := <-tr.Consume()
	// 		fmt.Println("Message: ", msg)
	// 	}
	// }()

	// if err := tr.ListenAndAccept(); err != nil {
	// 	log.Fatal(err)
	// }

	// select {}

	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: Onpeer func
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := server.FileServerOpts{
		ListenAddr:        ":3000",
		StorageRoot:       "3000_network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
	}

	s := server.NewFileServer(fileServerOpts)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
