package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

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
	fmt.Println("Hash: ", hash)
	hashStr := hex.EncodeToString(hash[:])
	fmt.Println("Hash String: ", hashStr)

	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	fmt.Println("sliceLen: ", sliceLen)
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
	pathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStream(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	pathKey := s.pathTransformFunc(key)
	_, err := os.Stat(pathKey.FullPath())
	if err == fs.ErrNotExist {
		return false
	}
	return true
}

func (s *Store) Delete(key string) error {
	pathKey := s.pathTransformFunc(key)
	defer func() {
		log.Println("Deleted: <", pathKey.FullPath(), "> from disk")
	}()

	err := os.RemoveAll(pathKey.FirstPathName())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.pathTransformFunc(key)

	return os.Open(pathKey.FullPath())
}

func (s *Store) writeStream(r io.Reader, key string) error {
	pathkey := s.pathTransformFunc(key)
	if err := os.MkdirAll(pathkey.PathName, os.ModePerm); err != nil {
		return err
	}

	pathAndFilename := pathkey.FullPath()

	f, err := os.Create(pathAndFilename)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("Written (%d) bytes to disk: %s\n", n, pathAndFilename)

	return nil
}
