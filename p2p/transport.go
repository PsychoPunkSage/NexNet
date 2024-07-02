package p2p

// Peer: Representation of remote node.
type Peer interface {
	Close() error
}

// Transport: Anything that handle communication between node in the Network.
// This can be of form TCP, UDP, websockets, etc.
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
