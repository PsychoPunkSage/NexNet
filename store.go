package main

import (
	"io"
	"log"
	"os"
)

type PathTransformFunc func(string) string

var DefaultPathTransformFunc = func(key string) string {
	return key
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

	filename := "Somefilename"

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
