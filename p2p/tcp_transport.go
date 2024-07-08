package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer: represents a node over TCP connection.
type TCPPeer struct {
	// Underlying connection of the peer. TCP connection (here)
	net.Conn
	/*
		dial & retrieve a conn : outbound == true
		accept & retrieve a conn : outbound == false
	*/
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(data []byte) error {
	_, err := p.Conn.Write(data)
	return err
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

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on %s\n", t.ListenAddr)

	return nil
}

func (t *TCPTransport) ListenAddress() string {
	return t.ListenAddr
}

// Consume : return read-only channel for reading messages received from other Peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	// To implement read-only interface.
	return t.rpcch
}

// Close: implements transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial: implements transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			// To Stop listening Once net-commection has been closed.
			return
		}

		if err != nil {
			fmt.Println("TCP accept error:", err)
		}

		fmt.Printf("new incoming connection: %+v\n", conn)
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Println("Dropping Peer Connection:", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

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
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		// fmt.Println(reflect.TypeOf(err))
		// panic(err)
		if err == net.ErrClosed {
			return
		}

		rpc.From = conn.RemoteAddr()

		if rpc.Stream {
			peer.Wg.Add(1)
			fmt.Printf("[%s] Incoming Stream, waiting...\n", rpc.From)
			peer.Wg.Wait()
			fmt.Printf("[%s] Stream Closed,resuming read loop\n", rpc.From)
			fmt.Println("=============================================================")
			continue
		}

		// if err != nil {
		// 	fmt.Println("TCP READ Error:", err)
		// 	continue
		// }

		t.rpcch <- rpc // pass the received RPC message to another part of the program for further processing.
	}
}
