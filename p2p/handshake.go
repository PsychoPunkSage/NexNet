package p2p

import "errors"

// ErrInvalidHandShake : returned if handshake between local and remote node couldn't be established.
var ErrInvalidHandShake = errors.New("invalid handshake")

// HandshakeFunc:
type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error { return nil }
