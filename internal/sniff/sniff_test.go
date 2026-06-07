package sniff

import "testing"

// TestReadMpeg4FtypBoxShortHeader covers headers between 8 and 15 bytes, which
// previously panicked when reading header[8:16].
func TestReadMpeg4FtypBoxShortHeader(t *testing.T) {
	for _, n := range []int{0, 4, 8, 12, 15} {
		header := make([]byte, n)
		if n >= 8 {
			copy(header[4:8], "ftyp")
		}
		if _, ok := AudioExtension(header); ok {
			t.Errorf("AudioExtension(%d-byte header) unexpectedly matched", n)
		}
	}
}

// TestAudioExtensionFtyp confirms a well-formed ftyp box (with a compatible
// brand right at the end of the header) is recognised. The exact extension may
// be .m4a or .mp4 since both brands match an ftyp box.
func TestAudioExtensionFtyp(t *testing.T) {
	header := []byte{
		0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p',
		'M', '4', 'A', ' ', 0x00, 0x00, 0x00, 0x00,
		'M', '4', 'A', ' ', 'm', 'p', '4', '2',
	}
	ext, ok := AudioExtension(header)
	if !ok || (ext != ".m4a" && ext != ".mp4") {
		t.Errorf("AudioExtension(ftyp) = (%q, %v), want a non-empty mp4/m4a extension", ext, ok)
	}
}

func TestImageExtension(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
		want   string
		ok     bool
	}{
		{"jpeg returns .jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0}, ".jpg", true},
		{"png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}, ".png", true},
		{"gif", []byte("GIF89a"), ".gif", true},
		{"bmp", []byte("BM     "), ".bmp", true},
		{"webp", []byte("RIFF\x00\x00\x00\x00WEBPVP8 "), ".webp", true},
		{"wav is not webp", []byte("RIFF\x00\x00\x00\x00WAVEfmt "), "", false},
		{"unknown", []byte("xxxxxxxx"), "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ImageExtension(tt.header)
			if ok != tt.ok || got != tt.want {
				t.Errorf("ImageExtension = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}
