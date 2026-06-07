package ximalaya

import (
	_ "embed"
	"encoding/binary"
)

var x3mKey = [...]byte{
	'3', '9', '8', '9', 'd', '1', '1', '1',
	'a', 'a', 'd', '5', '6', '1', '3', '9',
	'4', '0', 'f', '4', 'f', 'c', '4', '4',
	'b', '6', '3', '9', 'b', '2', '9', '2',
}

const x3mHeaderSize = 1024

var x3mScrambleTable = [x3mHeaderSize]uint16{}

//go:embed x3m_scramble_table.bin
var x3mScrambleTableBytes []byte

func init() {
	if len(x3mScrambleTableBytes) != 2*x3mHeaderSize {
		panic("invalid x3m scramble table")
	}
	for i := range x3mScrambleTable {
		x3mScrambleTable[i] = binary.LittleEndian.Uint16(x3mScrambleTableBytes[i*2:])
	}
}

// decryptX3MHeader decrypts the header of ximalaya .x3m file.
// make sure input src is 1024 (x3mHeaderSize) bytes long.
func decryptX3MHeader(src []byte) []byte {
	dst := make([]byte, len(src))
	// The scramble table is sized for a full header; bound both indices so a
	// short or oversized buffer can't cause an out-of-range access.
	n := min(len(src), x3mHeaderSize)
	for dstIdx := range n {
		srcIdx := int(x3mScrambleTable[dstIdx])
		if srcIdx < len(src) {
			dst[dstIdx] = src[srcIdx] ^ x3mKey[dstIdx%len(x3mKey)]
		}
	}
	return dst
}
