package ncm

import (
	"bytes"
	"encoding/binary"
	"testing"

	"git.um-react.app/um/cli/algo/common"
	"go.uber.org/zap"
)

func newTestDecoder(data []byte) common.Decoder {
	return NewDecoder(&common.DecoderParams{
		Reader: bytes.NewReader(data),
		Logger: zap.NewNop(),
	})
}

// TestNCMValidateRejectsCraftedInput feeds malformed .ncm inputs to Validate and
// requires an error rather than a panic (out-of-range slice, 4GB allocation, etc.).
func TestNCMValidateRejectsCraftedInput(t *testing.T) {
	magic := []byte("CTENFDAM")

	craftKeyLen := func(n uint32) []byte {
		var b bytes.Buffer
		b.Write(magic)
		b.Write([]byte{0, 0}) // 2-byte gap
		_ = binary.Write(&b, binary.LittleEndian, n)
		return b.Bytes()
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"short magic", []byte("CTE")},
		{"wrong magic", append([]byte("XXXXXXXX"), make([]byte, 32)...)},
		{"magic only", magic},
		{"huge key length", craftKeyLen(0xFFFFFFFF)},
		{"key length beyond file", craftKeyLen(1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoder(tt.data)
			if err := d.Validate(); err == nil {
				t.Error("expected an error for crafted input, got nil")
			}
		})
	}
}
