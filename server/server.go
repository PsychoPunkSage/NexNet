package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/PsychoPunkSage/NexNet/cryptography"
	"github.com/PsychoPunkSage/NexNet/p2p"
	store "github.com/PsychoPunkSage/NexNet/storage"
)

const PrependSig int64 = 16

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string
	Key string
}

type MessageDeleteFile struct {
	ID  string
	Key string
}

type FileServerOpts struct {
	ID                string
	EncKey            []byte
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

	if len(opts.ID) == 0 {
		opts.ID = cryptography.GenerateId()
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          store.NewStream(storeOpts),
		quitCh:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(s.ID, key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.ListenAddress(), key)
		_, r, err := s.store.Read(s.ID, key)
		return r, err
	}

	fmt.Printf("[%s] Don't have file (%s) locally, fetching from network...\n", s.Transport.ListenAddress(), key)

	msg := Message{
		Payload: MessageGetFile{
			ID:  s.ID,
			Key: cryptography.HashKey(key),
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		fmt.Println("receiving stream from peer:", peer.RemoteAddr())
		// fileBuf := new(bytes.Buffer)
		// n, err := io.CopyN(fileBuf, peer, 22)
		// if err != nil {
		// 	return nil, err
		// }

		var size int64
		binary.Read(peer, binary.LittleEndian, &size)
		// To Store Incoming File in the Calling Network.
		n, err := s.store.WriteDecrypt(s.EncKey, io.LimitReader(peer, size), s.ID, key)
		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] Received (%d) bytes ove the network from <%s>\n", s.Transport.ListenAddress(), n, peer.RemoteAddr())
		peer.CloseStream()
	}

	_, r, err := s.store.Read(s.ID, key)
	return r, err
}

func (s *FileServer) Remove(key string) error {
	if !s.store.Has(s.ID, key) {
		fmt.Printf("[%s] The file (%s) is not present in the disk.\n", s.Transport.ListenAddress(), key)
		return fmt.Errorf("file is not present")
	}

	fmt.Printf("[%s] File (%s) found locally, deleting it...\n", s.Transport.ListenAddress(), key)

	// Message to be broadcasted.
	msg := Message{
		Payload: MessageDeleteFile{
			ID:  s.ID,
			Key: cryptography.HashKey(key),
		},
	}

	// Broadcasting the message
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		fmt.Printf("[%s] receiving DELETE stream from peer: [%s]", s.Transport.ListenAddress(), peer.RemoteAddr())

		if err := s.store.Delete(s.ID, key); err != nil {
			return err
		}
	}

	fmt.Printf("[%s] File (%s) DELETED.\n", s.Transport.ListenAddress(), key)
	return s.store.Delete(s.ID, key)
}

func (s *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	n, err := s.store.Write(tee, s.ID, key)
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
			ID:   s.ID,
			Key:  cryptography.HashKey(key),
			Size: n + PrependSig,
		},
	}

	// Broadcast the FileKey and FileSize to be stored.
	if err = s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(5 * time.Millisecond)

	// payload := []byte("VERY LARGE FILE CONTENT")
	////// USE multiwriter here.
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	nn, err := cryptography.CopyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] recv & written (%d) bytes to disk\n", s.Transport.ListenAddress(), nn)

	// for _, peer := range s.peers {
	// 	// if err := peer.Send(payload); err != nil {
	// 	// 	return err
	// 	// }
	// 	peer.Send([]byte{p2p.IncomingStream}) // About to Stream FileStorage Payload.
	// 	n, err := cryptography.CopyEncrypt(s.EncKey, fileBuffer, peer)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
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
				log.Println("decodeing err:", err)
			}

			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Println("handleMessage err:", err)
				// return
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
			fmt.Printf("[%s] attempting to connect with <%s>\n", s.Transport.ListenAddress(), addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error: ", err)
			}
		}(addr)
	}

	return nil
}

// func (s *FileServer) stream(msg *Message) error {
// 	peers := []io.Writer{}
// 	for _, peer := range s.peers {
// 		peers = append(peers, peer)
// 	}

// 	multiWriter := io.MultiWriter(peers...)
// 	return gob.NewEncoder(multiWriter).Encode(msg)
// }

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage}) // Cause we are sending message to all the peers.
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch t := msg.Payload.(type) {
	case *MessageStoreFile:
		fmt.Println("Received MessageStoreFile")
		return s.handleMessageStoreFile(from, t)

	case *MessageGetFile:
		fmt.Println("Received MessageGetFile")
		return s.handleMessageGetFile(from, t)

	case *MessageDeleteFile:
		fmt.Println("Received MessageDeleteFile")
		return s.handleMessageDeleteFile(from, t)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg *MessageStoreFile) error {
	fmt.Printf("Received Message: %v\n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer {%s} not found", from)
	}

	n, err := s.store.Write(io.LimitReader(peer, msg.Size), msg.ID, msg.Key)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Written (%d) bytes to disk\n", s.Transport.ListenAddress(), n)

	peer.CloseStream() // Streaming is OVER!!

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg *MessageGetFile) error {
	fmt.Printf("Received Message: %v\n", msg)

	if !s.store.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] file (%s) not found", s.Transport.ListenAddress(), msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.ListenAddress(), msg.Key)

	size, r, err := s.store.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Printf("[%s] Closing ReadCloser\n", s.Transport.ListenAddress())
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer {%s} not found", from)
	}

	// for `peepBuf` to know that a Stream is comming
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, size)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written (%d) bytes over the network to <%s>\n", s.Transport.ListenAddress(), n, from)
	return nil
}

func (s *FileServer) handleMessageDeleteFile(from string, msg *MessageDeleteFile) error {
	fmt.Printf("Received Message: %v\n", msg)

	if !s.store.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] file (%s) not found", s.Transport.ListenAddress(), msg.Key)
	}

	_, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer {%s} not found", from)
	}

	if err := s.store.Delete(msg.ID, msg.Key); err != nil {
		return err
	}

	fmt.Printf("[%s] Deleted file (%s) from disk\n", s.Transport.ListenAddress(), msg.Key)
	return nil
}

func init() {
	gob.Register(&MessageStoreFile{})
	gob.Register(&MessageGetFile{})
	gob.Register(&MessageDeleteFile{})
}
