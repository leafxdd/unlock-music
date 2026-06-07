package utils

import (
	"bytes"
	"crypto/aes"
	"testing"
)

func TestPKCS7UnPadding(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		want    []byte
		wantErr bool
	}{
		{"valid pad 4", []byte{1, 2, 3, 4, 4, 4, 4, 4}, []byte{1, 2, 3, 4}, false},
		{"valid pad 1", []byte{9, 1}, []byte{9}, false},
		{"whole block is pad", []byte{8, 8, 8, 8, 8, 8, 8, 8}, []byte{}, false},
		{"empty input", []byte{}, nil, true},
		{"pad byte zero", []byte{1, 2, 0}, nil, true},
		{"pad exceeds length", []byte{1, 2, 9}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PKCS7UnPadding(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !bytes.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecryptAES128ECBRejectsBadInput(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes
	tests := []struct {
		name    string
		data    []byte
		key     []byte
		wantErr bool
	}{
		{"valid one block", make([]byte, 16), key, false},
		{"valid two blocks", make([]byte, 32), key, false},
		{"empty data", []byte{}, key, true},
		{"misaligned data", make([]byte, 17), key, true},
		{"short key", make([]byte, 16), []byte("short"), true},
		{"long key", make([]byte, 16), make([]byte, 32), true},
		{"nil key", make([]byte, 16), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptAES128ECB(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != len(tt.data) {
				t.Errorf("got %d bytes, want %d", len(got), len(tt.data))
			}
		})
	}
}

// TestDecryptAES128ECBRoundTrip confirms the hardening did not change the
// decryption result for valid input.
func TestDecryptAES128ECBRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef")
	plain := []byte("sixteen byte msg") // exactly one block
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	ct := make([]byte, len(plain))
	block.Encrypt(ct, plain)

	got, err := DecryptAES128ECB(ct, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Errorf("round trip failed: got %q want %q", got, plain)
	}
}
