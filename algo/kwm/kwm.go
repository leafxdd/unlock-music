package kwm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"git.um-react.app/um/cli/algo/common"
)

const magicHeader1 = "yeelion-kuwo-tme"
const magicHeader2 = "yeelion-kuwo\x00\x00\x00\x00"
const keyPreDefined = "MoOtOiTvINGwd2E6n0E1i7L5t2IoOoNk"

type Decoder struct {
	rd io.ReadSeeker

	cipher common.StreamDecoder
	offset int

	outputExt string
	bitrate   int
}

func (d *Decoder) GetAudioExt() string {
	return "." + d.outputExt
}

func NewDecoder(p *common.DecoderParams) common.Decoder {
	return &Decoder{rd: p.Reader}
}

// Validate checks if the file is a valid Kuwo .kw file.
// rd will be seeked to the beginning of the encrypted audio.
func (d *Decoder) Validate() error {
	header := make([]byte, 0x400) // kwm header is fixed to 1024 bytes
	_, err := io.ReadFull(d.rd, header)
	if err != nil {
		return fmt.Errorf("kwm read header: %w", err)
	}

	// check magic header, 0x00 - 0x0F
	magicHeader := header[:0x10]
	if !bytes.Equal([]byte(magicHeader1), magicHeader) &&
		!bytes.Equal([]byte(magicHeader2), magicHeader) {
		return errors.New("kwm magic header not matched")
	}

	d.cipher = newKwmCipher(header[0x18:0x20])                      // Crypto Key, 0x18 - 0x1F
	d.bitrate, d.outputExt = parseBitrateAndType(header[0x30:0x38]) // Bitrate & File Extension, 0x30 - 0x38

	return nil
}

func parseBitrateAndType(header []byte) (int, string) {
	tmp := strings.TrimRight(string(header), "\x00")
	sep := strings.IndexFunc(tmp, func(r rune) bool {
		return !unicode.IsDigit(r)
	})
	if sep < 0 {
		// No non-digit separator (all digits or empty): treat the whole field as
		// the bitrate with no extension, rather than slicing with sep == -1.
		bitrate, _ := strconv.Atoi(tmp)
		return bitrate, ""
	}

	bitrate, _ := strconv.Atoi(tmp[:sep]) // just ignore the error
	outputExt := strings.ToLower(tmp[sep:])
	return bitrate, outputExt
}

func (d *Decoder) Read(b []byte) (int, error) {
	n, err := d.rd.Read(b)
	if n > 0 {
		d.cipher.Decrypt(b[:n], d.offset)
		d.offset += n
	}
	return n, err
}

func padOrTruncate(raw string, length int) string {
	lenRaw := len(raw)
	if lenRaw == 0 {
		return string(make([]byte, length))
	}
	if lenRaw >= length {
		return raw[:length]
	}
	out := make([]byte, length)
	for i := range out {
		out[i] = raw[i%lenRaw]
	}
	return string(out)
}

func init() {
	// Kuwo Mp3/Flac
	common.RegisterDecoder("kwm", false, NewDecoder)
}
