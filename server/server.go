package server

import (
	"fmt"
	"io"
	"log"

	"github.com/PsychoPunkSage/NexNet/p2p"
	store "github.com/PsychoPunkSage/NexNet/storage"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc store.PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

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
	}
}

func (s *FileServer) Store(key string, r io.Reader) error {
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

func (s *FileServer) loop() {
	defer func() {
		fmt.Println("File Server stopped due to user Quit action.")
		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
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
