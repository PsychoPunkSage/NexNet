package cryptography

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "AP is here"
	src := bytes.NewReader([]byte(payload))
	fmt.Println("src:>", src)
	dst := new(bytes.Buffer)
	key := NewEncryptionKey()
	fmt.Println("key:>", key)

	_, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	// fmt.Println("dst:>", dst.String())

	out := new(bytes.Buffer)
	nn, err := CopyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	// fmt.Println("out:>", out.String())

	assert.Equal(t, out.String(), payload)
	assert.Equal(t, nn, 26) // 26 = 16(iv) + 10(payload)
}
