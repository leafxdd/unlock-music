package pc_kugou_db

import (
	"reflect"
	"testing"
)

func TestDerivePageAESKey_Page0(t *testing.T) {
	expectedKey := []byte{0x19, 0x62, 0xc0, 0x5f, 0xa2, 0xeb, 0xbe, 0x24, 0x28, 0xff, 0x52, 0x2b, 0x9e, 0x03, 0xea, 0xd4}
	pageKey := derivePageKey(0)
	if !reflect.DeepEqual(expectedKey, pageKey) {
		t.Errorf("Derived AES key for page 0 does not match expected value: got %v, want %v", pageKey, expectedKey)
	}
}

func TestDerivePageAESIv_Page0(t *testing.T) {
	expectedIv := []byte{0x05, 0x5a, 0x67, 0x35, 0x93, 0x89, 0x2d, 0xdf, 0x3a, 0xb3, 0xb3, 0xc6, 0x21, 0xc3, 0x48, 0x02}
	pageKey := derivePageIv(0)
	if !reflect.DeepEqual(expectedIv, pageKey) {
		t.Errorf("Derived AES iv for page 0 does not match expected value: got %v, want %v", pageKey, expectedIv)
	}
}

// TestDecryptDatabaseRejectsTinyBuffer covers the buffers shorter than the
// SQLite header, which previously panicked on the buffer[:16] slice.
func TestDecryptDatabaseRejectsTinyBuffer(t *testing.T) {
	for _, size := range []int{0, 1, 8, 15} {
		if err := decryptDatabase(make([]byte, size)); err == nil {
			t.Errorf("decryptDatabase(%d bytes) = nil, want error", size)
		}
	}
}

// TestDecryptDatabasePlainSQLitePassthrough confirms a buffer that already
// carries the SQLite header is treated as an unencrypted database.
func TestDecryptDatabasePlainSQLitePassthrough(t *testing.T) {
	buf := make([]byte, PAGE_SIZE)
	copy(buf, SQLITE_HEADER)
	if err := decryptDatabase(buf); err != nil {
		t.Errorf("decryptDatabase(plain sqlite) = %v, want nil", err)
	}
}
