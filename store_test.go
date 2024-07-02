package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStream(opts)
	key := "myspecialphotos"
	data := []byte("testing the Store withStream func")

	//  Create
	if err := store.writeStream(bytes.NewReader(data), key); err != nil {
		t.Error(err)
	}

	if ok := store.Has(key); !ok {
		t.Errorf("Wrong key Passed")
	}

	// Read
	r, err := store.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("want %s got %s", string(b), string(data))
	}

	// Delete
	store.Delete(key)
}

func TestStoreStream(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStream(opts)
	key := "myspecialphotos"
	data := []byte("testing the Store withStream func")

	//  Create
	if err := store.writeStream(bytes.NewReader(data), key); err != nil {
		t.Error(err)
	}
}

func TestReadStream(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStream(opts)
	key := "myspecialphotos"
	data := []byte("testing the Store withStream func")

	// Create
	if err := store.writeStream(bytes.NewReader(data), key); err != nil {
		t.Error(err)
	}

	// Read
	r, err := store.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("want %s got %s", string(b), string(data))
	}
}

func TestDeleteStream(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStream(opts)
	key := "myspecialphotos"
	data := []byte("testing the Store withStream func")

	// Create
	if err := store.writeStream(bytes.NewReader(data), key); err != nil {
		t.Error(err)
	}

	err := store.Delete(key)
	if err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestPathTransformFunc(t *testing.T) {
	key := "mybestpic"
	pathkey := CASPathTransformFunc(key)
	// fmt.Println(pathname)
	expectedPathname := "1b150/aae86/eedae/268f6/589f4/0fb48/b2a0d/47ff4"
	if pathkey.PathName != expectedPathname {
		t.Error(t, "Want %s Got %s", expectedPathname, pathkey.PathName)
	}

	expectedOriginal := "1b150aae86eedae268f6589f40fb48b2a0d47ff4"
	if pathkey.Filename != expectedOriginal {
		t.Errorf("Want %s Got %s", expectedOriginal, pathkey.Filename)
	}
}
