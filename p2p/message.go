package p2p

import "net"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC : arbitrary data that is being sent over each transport between 2 node (peer) in the network
type RPC struct {
	Payload []byte
	From    net.Addr
	Stream  bool
}
