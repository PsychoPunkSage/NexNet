package main

import (
	"fmt"
	"log"

	"github.com/PsychoPunkSage/NexNet/p2p"
)

func main() {
	tr := p2p.NewTCPTransport(":3000")

	fmt.Println("AP is here..")

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
