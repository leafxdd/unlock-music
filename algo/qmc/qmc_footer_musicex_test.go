package qmc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestNewMusicExTagRejectsCraftedInput verifies the musicex footer parser rejects
// malformed footers (bad magic/version, oversized TagSize) with an error instead
// of panicking or attempting a huge allocation.
func TestNewMusicExTagRejectsCraftedInput(t *testing.T) {
	build := func(tagSize, version uint32, magic string) []byte {
		b := make([]byte, 16)
		binary.LittleEndian.PutUint32(b[0:4], tagSize)
		binary.LittleEndian.PutUint32(b[4:8], version)
		copy(b[8:], magic)
		return b
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"too short to seek", make([]byte, 8)},
		{"bad magic", build(0xC0, 1, "NOTMAGIC")},
		{"oversized tag size", build(0xFFFFFFFF, 1, "musicex\x00")},
		{"tag size exceeds file", build(0xC0, 1, "musicex\x00")},
		{"bad version", build(0xC0, 2, "musicex\x00")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewMusicExTag(bytes.NewReader(tt.data)); err == nil {
				t.Error("expected an error for crafted footer, got nil")
			}
		})
	}
}
