package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: DefaultPathTransformFunc,
	}
	store := NewStream(opts)
	assert.Equal(t, store.pathTransformFunc("AP"), "AP")

	data := bytes.NewReader([]byte("some png data"))
	if err := store.writeStream(data, "myspecialpic"); err != nil {
		t.Error(err)
	}
}
