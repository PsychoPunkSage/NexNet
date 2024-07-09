package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "PPSNetwork"

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p *PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

type PathTransformFunc func(string) PathKey

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	// fmt.Println("Hash: ", hash)
	hashStr := hex.EncodeToString(hash[:])
	// fmt.Println("Hash String: ", hashStr)

	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	// fmt.Println("sliceLen: ", sliceLen)
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, i*blocksize+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type StoreOpts struct {
	// Folder name of Root; Contains all the folders/files of the system.
	Root              string
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStream(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Println("Deleted: <", pathKey.FullPath(), "> from disk")
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())

	err := os.RemoveAll(firstPathNameWithRoot)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

func (s *Store) Write(r io.Reader, key string) (int64, error) {
	return s.writeStream(r, key)
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	file, err := os.Open(fullPathKeyWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil
}

func (s *Store) writeStream(r io.Reader, key string) (int64, error) {
	pathkey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.FullPath())

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0, err
	}
	// When we read from a connection, the conn will not always return a file.
	// Basically, storage keeps on waiting for new stuffs
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}

	return n, nil
}
