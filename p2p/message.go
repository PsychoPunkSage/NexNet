package p2p

// Message : arbitrary data that is being sent over each transport between 2 node (peer) in the network
type Message struct {
	Payload []byte
}
