package utils

import (
	"crypto/aes"
	"errors"
	"fmt"
)

// PKCS7UnPadding removes PKCS#7 padding. It validates the trailing pad length
// so crafted input cannot produce a negative or out-of-range slice (which would
// otherwise panic).
func PKCS7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: empty input")
	}
	unPadding := int(data[length-1])
	if unPadding == 0 || unPadding > length {
		return nil, fmt.Errorf("pkcs7: invalid padding length %d for %d bytes of data", unPadding, length)
	}
	return data[:length-unPadding], nil
}

// DecryptAES128ECB decrypts data with AES-128 in ECB mode. The key must be
// 16 bytes and data a non-empty multiple of the block size; otherwise it returns
// an error rather than panicking on a nil cipher or a misaligned block slice.
func DecryptAES128ECB(data, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("aes-128-ecb: key must be 16 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes-128-ecb: new cipher: %w", err)
	}
	const size = aes.BlockSize // 16
	if len(data) == 0 || len(data)%size != 0 {
		return nil, fmt.Errorf("aes-128-ecb: data length %d is not a positive multiple of block size %d", len(data), size)
	}
	decrypted := make([]byte, len(data))
	for bs := 0; bs < len(data); bs += size {
		block.Decrypt(decrypted[bs:bs+size], data[bs:bs+size])
	}
	return decrypted, nil
}
