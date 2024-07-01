package main

import (
	"bytes"
	"testing"
)

func TestStore(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: CASPathTransformFunc,
	}
	store := NewStream(opts)

	data := bytes.NewReader([]byte("testing the Store withStream func"))
	if err := store.writeStream(data, "myspecialpic"); err != nil {
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
	if pathkey.Original != expectedOriginal {
		t.Errorf("Want %s Got %s", expectedOriginal, pathkey.Original)
	}
}
