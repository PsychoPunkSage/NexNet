package p2p

import (
	"fmt"
	"net"
)

// TCPPeer: represents a node over TCP connection.
type TCPPeer struct {
	// conn: Underlying connection of th peer.
	conn net.Conn
	/*
		dial & retrieve a conn : outbound == true
		accept & retrieve a conn : outbound == false
	*/
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPPeer) Close() error {
	// To implement Peer interface
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume : return read-only channel for reading messages received from other Peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	// To implement read-only interface.
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("TCP accept error:", err)
		}

		fmt.Printf("new incoming connection: %+v\n", conn)
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		fmt.Println("Dropping Peer Connection:", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err = t.HandshakeFunc(peer); err != nil {
		fmt.Println("TCP Handshake Error:", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			// fmt.Println("TCP OnPeer Error:", err)
			return
		}
	}

	// Read Loop
	rpc := RPC{}
	for {
		err = t.Decoder.Decode(conn, &rpc)
		// fmt.Println(reflect.TypeOf(err))
		// panic(err)
		if err == net.ErrClosed {
			return
		}

		if err != nil {
			fmt.Println("TCP READ Error:", err)
			continue
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc // pass the received RPC message to another part of the program for further processing.
	}
}
