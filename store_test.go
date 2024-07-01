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

	data := bytes.NewReader([]byte("some png data"))
	if err := store.writeStream(data, "myspecialpic"); err != nil {
		t.Error(err)
	}
}

func TestPathTransformFunc(t *testing.T) {
	key := "mybestpic"
	pathname := CASPathTransformFunc(key)
	// fmt.Println(pathname)
	expectedPathname := "1b150/aae86/eedae/268f6/589f4/0fb48/b2a0d/47ff4"
	if pathname != expectedPathname {
		t.Error(t, "Want %s Got %s", expectedPathname, pathname)
	}
}
