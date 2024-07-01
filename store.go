package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(string) string

var DefaultPathTransformFunc = func(key string) string {
	return key
}

func CASPathTransformFunc(key string) string {
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

	return strings.Join(paths, "/")
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

func (s *Store) writeStream(r io.Reader, key string) error {
	pathName := s.pathTransformFunc(key)
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	filenameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameBytes[:])
	pathAndFilename := pathName + "/" + filename

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
