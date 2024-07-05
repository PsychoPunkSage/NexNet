package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/PsychoPunkSage/NexNet/p2p"
	store "github.com/PsychoPunkSage/NexNet/storage"
)

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc store.PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *store.Store
	quitCh chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := store.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          store.NewStream(storeOpts),
		quitCh:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// // 1. Store the file in the disk
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	n, err := s.store.Write(tee, key)
	if err != nil {
		return err
	}

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// fmt.Println("Buffer:>", buf.Bytes())
	// fmt.Printf("DataMessage:> %v\n", p)

	// return s.broadcast(&Message{
	// 	From:    s.Transport.ListenAddress(),
	// 	Payload: p,
	// })

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: n,
		},
	}

	msgBuf := new(bytes.Buffer)

	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 3)

	// payload := []byte("VERY LARGE FILE CONTENT")
	for _, peer := range s.peers {
		// if err := peer.Send(payload); err != nil {
		// 	return err
		// }
		n, err := io.Copy(peer, buf)
		if err != nil {
			return err
		}
		fmt.Println("recv & written bytes to disk: ", n)
	}

	return nil
}

func (s *FileServer) Store(key string, r io.Reader) (int64, error) {
	return s.store.Write(r, key)
}

func (s *FileServer) Stop() {
	close(s.quitCh)
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Println("Connected with remote Peer:", p.RemoteAddr().String())
	return nil
}

func (s *FileServer) loop() {
	defer func() {
		fmt.Println("File Server stopped due to user Quit action.")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
			}

			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Println(err)
				return
			}

			// fmt.Printf("recv: %s\n", string(msg.Payload.([]byte)))
			// fmt.Printf("Payload: %v\n", msg.Payload)

			// peer, ok := s.peers[rpc.From.String()]
			// if !ok {
			// 	panic("peer not found in peers map")
			// }

			// fmt.Println("Peer: ", peer.RemoteAddr())

			// b := make([]byte, 1000)
			// if _, err := peer.Read(b); err != nil {
			// 	panic(err)
			// }

			// fmt.Printf("Data: %s\n", string(b))

			// peer.(*p2p.TCPPeer).Wg.Done()
			// if err := s.handleMessage(&m); err != nil {
			// 	log.Fatal(err)
			// }
		case <-s.quitCh:
			return
		}
	}
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			// In case of empty string... SKIP
			continue
		}

		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error: ", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	multiWriter := io.MultiWriter(peers...)
	return gob.NewEncoder(multiWriter).Encode(msg)
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch t := msg.Payload.(type) {
	case *MessageStoreFile:
		fmt.Printf("Received MessageStoreFile")
		return s.handleMessageStoreFile(from, t)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg *MessageStoreFile) error {
	fmt.Printf("Received Message: %v\n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer {%s} not found", from)
	}

	_, err := s.store.Write(io.LimitReader(peer, msg.Size), msg.Key)
	if err != nil {
		return err
	}

	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
}

func init() {
	gob.Register(&MessageStoreFile{})
}
