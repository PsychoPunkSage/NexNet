package p2p

import (
	"net"
)

// Peer: Representation of remote node.
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Transport: Anything that handle communication between node in the Network.
// This can be of form TCP, UDP, websockets, etc.
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	ListenAddress() string
}
