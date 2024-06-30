package main

import (
	"fmt"
	"log"

	"github.com/PsychoPunkSage/NexNet/p2p"
)

func OnPeer(p2p.Peer) error {
	fmt.Println("Doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	fmt.Println("AP is here..")

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Println("Message: ", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
