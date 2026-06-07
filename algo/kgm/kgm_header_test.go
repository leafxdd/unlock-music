package kgm

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestKgmHeaderRejectsCraftedAudioHashLen verifies the v5 header parser bounds the
// attacker-controlled audio-hash length by the remaining file instead of trying a
// huge allocation.
func TestKgmHeaderRejectsCraftedAudioHashLen(t *testing.T) {
	var b bytes.Buffer
	b.Write(kgmHeader)                                       // 16-byte magic
	_ = binary.Write(&b, binary.LittleEndian, uint32(0x30))  // AudioOffset
	_ = binary.Write(&b, binary.LittleEndian, uint32(5))     // CryptoVersion = 5
	_ = binary.Write(&b, binary.LittleEndian, uint32(0))     // CryptoSlot
	b.Write(make([]byte, 16))                                // CryptoTestData
	b.Write(make([]byte, 16))                                // CryptoKey
	b.Write(make([]byte, 8))                                 // v5 gap
	_ = binary.Write(&b, binary.LittleEndian, uint32(0xFFFFFFFF)) // audioHashLen

	h := &header{}
	if err := h.FromBytes(bytes.NewReader(b.Bytes())); err == nil {
		t.Error("expected an error for oversized audio hash length, got nil")
	}
}

// TestKgmHeaderRejectsShortInput confirms short/empty headers error rather than panic.
func TestKgmHeaderRejectsShortInput(t *testing.T) {
	for _, n := range []int{0, 8, 16, 40} {
		h := &header{}
		if err := h.FromBytes(bytes.NewReader(make([]byte, n))); err == nil {
			t.Errorf("expected an error for %d-byte header, got nil", n)
		}
	}
}
