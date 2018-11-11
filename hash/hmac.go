package hash

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"hash"
)

type HMAC struct {
	hmac hash.Hash
}

func NewHMAC(key string) HMAC {
	h := hmac.New(sha1.New, []byte(key))
	return HMAC{
		hmac: h,
	}
}

func (h HMAC) Hash(input string) []byte {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	return h.hmac.Sum(nil)
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
