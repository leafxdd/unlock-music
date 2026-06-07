package sniff

import "bytes"

// ref: https://mimesniff.spec.whatwg.org
var imageMIMEs = map[string]Sniffer{
	"image/jpeg": prefixSniffer{0xFF, 0xD8, 0xFF},
	"image/png":  prefixSniffer{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'},
	"image/bmp":  prefixSniffer("BM"),
	"image/webp": webpSniffer{},
	"image/gif":  prefixSniffer("GIF8"),
}

// webpSniffer matches the "RIFF????WEBP" container signature. A bare "RIFF"
// prefix is ambiguous (WAV and AVI also start with RIFF), so the "WEBP" marker
// at offset 8 must be checked too.
type webpSniffer struct{}

func (webpSniffer) Sniff(header []byte) bool {
	return len(header) >= 12 &&
		bytes.HasPrefix(header, []byte("RIFF")) &&
		bytes.Equal(header[8:12], []byte("WEBP"))
}

// ImageMIME sniffs the well-known image types, and returns its MIME.
func ImageMIME(header []byte) (string, bool) {
	for ext, sniffer := range imageMIMEs {
		if sniffer.Sniff(header) {
			return ext, true
		}
	}
	return "", false
}

// ImageExtension is equivalent to ImageMIME, but returns file extension
func ImageExtension(header []byte) (string, bool) {
	mimeType, ok := ImageMIME(header)
	if !ok {
		return "", false
	}
	switch mimeType {
	case "image/jpeg":
		return ".jpg", true // prefer the conventional .jpg over .jpeg
	default:
		return "." + mimeType[6:], true // "image/" is 6 bytes
	}
}
