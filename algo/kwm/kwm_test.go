package kwm

import "testing"

// TestParseBitrateAndType covers the crafted-header cases that previously panicked
// when strings.IndexFunc returned -1 (an all-digit or empty bitrate/type field).
func TestParseBitrateAndType(t *testing.T) {
	tests := []struct {
		name        string
		header      []byte
		wantBitrate int
		wantExt     string
	}{
		{"normal flac", []byte("192flac\x00"), 192, "flac"},
		{"normal mp3", []byte("128mp3\x00\x00"), 128, "mp3"},
		{"all digits, no separator", []byte("128\x00\x00\x00\x00\x00"), 128, ""},
		{"all zero bytes", []byte{0, 0, 0, 0, 0, 0, 0, 0}, 0, ""},
		{"leading non-digit", []byte("ape\x00\x00\x00\x00\x00"), 0, "ape"},
		{"empty", []byte{}, 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBitrate, gotExt := parseBitrateAndType(tt.header)
			if gotBitrate != tt.wantBitrate || gotExt != tt.wantExt {
				t.Errorf("parseBitrateAndType(%q) = (%d, %q), want (%d, %q)",
					tt.header, gotBitrate, gotExt, tt.wantBitrate, tt.wantExt)
			}
		})
	}
}

// TestPadOrTruncate confirms the function now honours the requested length
// (previously it hard-coded 32 in the pad branch).
func TestPadOrTruncate(t *testing.T) {
	tests := []struct {
		raw    string
		length int
		want   string
	}{
		{"abc", 5, "abcab"},
		{"abcdef", 3, "abc"},
		{"xy", 4, "xyxy"},
		{"", 3, "\x00\x00\x00"},
		{"exact", 5, "exact"},
	}
	for _, tt := range tests {
		got := padOrTruncate(tt.raw, tt.length)
		if got != tt.want {
			t.Errorf("padOrTruncate(%q, %d) = %q, want %q", tt.raw, tt.length, got, tt.want)
		}
		if len(got) != tt.length {
			t.Errorf("padOrTruncate(%q, %d) length = %d, want %d", tt.raw, tt.length, len(got), tt.length)
		}
	}
}
