package hash

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
)

type HMAC struct {
	key []byte
}

func NewHMAC(key string) HMAC {
	return HMAC{
		key: []byte(key),
	}
}

func (h HMAC) Hash(input string) []byte {
	ha := hmac.New(sha1.New, h.key)
	ha.Write([]byte(input))
	return ha.Sum(nil)
}

func (h HMAC) DigestCheck(digest, input string) error {
	mac, err := hex.DecodeString(digest)
	if err != nil {
		return err
	}

	if hmac.Equal(mac, h.Hash(input)) {
		return nil
	}

	return errors.New("not equal digest")
}
