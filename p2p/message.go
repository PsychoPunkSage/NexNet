package p2p

import "net"

// RPC : arbitrary data that is being sent over each transport between 2 node (peer) in the network
type RPC struct {
	Payload []byte
	From    net.Addr
}
