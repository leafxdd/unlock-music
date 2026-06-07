package common

import (
	"testing"
)

func TestRegisterDecoder(t *testing.T) {
	origRegistry := DecoderRegistry
	defer func() { DecoderRegistry = origRegistry }()
	DecoderRegistry = nil

	RegisterDecoder("test", false, func(p *DecoderParams) Decoder { return nil })
	RegisterDecoder(".test2", true, func(p *DecoderParams) Decoder { return nil })

	if len(DecoderRegistry) != 2 {
		t.Fatalf("expected 2 decoders, got %d", len(DecoderRegistry))
	}
	if DecoderRegistry[0].Suffix != ".test" {
		t.Errorf("expected suffix .test, got %s", DecoderRegistry[0].Suffix)
	}
	if DecoderRegistry[1].Suffix != ".test2" {
		t.Errorf("expected suffix .test2, got %s", DecoderRegistry[1].Suffix)
	}
}

func TestGetDecoder(t *testing.T) {
	origRegistry := DecoderRegistry
	defer func() { DecoderRegistry = origRegistry }()
	DecoderRegistry = nil

	noopFactory := func(p *DecoderParams) Decoder { return nil }
	realFactory := func(p *DecoderParams) Decoder { return nil }

	RegisterDecoder("qmc0", false, realFactory)
	RegisterDecoder("mp3", true, noopFactory)
	RegisterDecoder("mflac", false, realFactory)

	tests := []struct {
		name      string
		filename  string
		skipNoop  bool
		wantCount int
	}{
		{"match encrypted ext", "song.qmc0", false, 1},
		{"match noop ext", "song.mp3", false, 1},
		{"skip noop", "song.mp3", true, 0},
		{"no match", "song.txt", false, 0},
		{"case insensitive", "SONG.QMC0", false, 1},
		{"with path", "/some/dir/song.mflac", false, 1},
		{"windows path", `C:\music\song.mflac`, false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDecoder(tt.filename, tt.skipNoop)
			if len(got) != tt.wantCount {
				t.Errorf("GetDecoder(%q, %v) returned %d decoders, want %d", tt.filename, tt.skipNoop, len(got), tt.wantCount)
			}
		})
	}
}

func TestQMCKeysGet(t *testing.T) {
	keys := QMCKeys{
		"song1": "key1",
		"song2": "key2",
	}

	val, ok := keys.Get("song1")
	if !ok || val != "key1" {
		t.Errorf("Get(song1) = (%q, %v), want (key1, true)", val, ok)
	}

	_, ok = keys.Get("nonexist")
	if ok {
		t.Error("Get(nonexist) should return false")
	}
}
