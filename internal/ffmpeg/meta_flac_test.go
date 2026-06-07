package ffmpeg

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"

	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	flac "github.com/go-flac/go-flac"
	"go.uber.org/zap"
)

type fakeMeta struct {
	title, album string
	artists      []string
}

func (m fakeMeta) GetTitle() string     { return m.title }
func (m fakeMeta) GetAlbum() string     { return m.album }
func (m fakeMeta) GetArtists() []string { return m.artists }

// tinyPNG returns the bytes of a valid 1x1 PNG. flacpicture decodes the image
// to read its dimensions, so the cover data must be a real image.
func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 1, G: 2, B: 3, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// buildFLAC constructs a minimal but parseable FLAC file with a StreamInfo
// block, a Vorbis comment (with the given fields) and, optionally, one cover.
func buildFLAC(t *testing.T, comments map[string]string, withCover bool) string {
	t.Helper()
	f := &flac.File{}
	f.Meta = append(f.Meta, &flac.MetaDataBlock{Type: flac.StreamInfo, Data: make([]byte, 34)})

	cmt := flacvorbis.New()
	for k, v := range comments {
		if err := cmt.Add(k, v); err != nil {
			t.Fatal(err)
		}
	}
	cb := cmt.Marshal()
	f.Meta = append(f.Meta, &cb)

	if withCover {
		pic, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "old", tinyPNG(t), "image/png")
		if err != nil {
			t.Fatal(err)
		}
		pb := pic.Marshal()
		f.Meta = append(f.Meta, &pb)
	}

	// readFLACStream requires a frame sync code; two bytes are enough to parse.
	f.Frames = []byte{0xFF, 0xF8}

	path := filepath.Join(t.TempDir(), "in.flac")
	if err := f.Save(path); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestUpdateMetaFlacPreservesTagsSingleCover is the regression test for the
// inverted error check (which dropped existing tags) and the duplicate-cover
// append. A pre-existing DATE tag must survive, TITLE must be overwritten, and
// the output must contain exactly one cover.
func TestUpdateMetaFlacPreservesTagsSingleCover(t *testing.T) {
	src := buildFLAC(t, map[string]string{"TITLE": "old title", "DATE": "2020"}, true)
	out := filepath.Join(t.TempDir(), "out.flac")

	params := &UpdateMetadataParams{
		Audio:       src,
		AudioExt:    ".flac",
		Meta:        fakeMeta{title: "new title", album: "new album", artists: []string{"artistA"}},
		AlbumArt:    tinyPNG(t),
		AlbumArtExt: ".png",
	}
	if err := updateMetaFlac(context.Background(), out, params, zap.NewNop()); err != nil {
		t.Fatal(err)
	}

	res, err := flac.ParseFile(out)
	if err != nil {
		t.Fatal(err)
	}

	covers := 0
	var comment *flacvorbis.MetaDataBlockVorbisComment
	for _, b := range res.Meta {
		switch b.Type {
		case flac.Picture:
			covers++
		case flac.VorbisComment:
			if c, err := flacvorbis.ParseFromMetaDataBlock(*b); err == nil {
				comment = c
			}
		}
	}

	if covers != 1 {
		t.Errorf("expected exactly 1 cover block, got %d", covers)
	}
	if comment == nil {
		t.Fatal("no vorbis comment in output")
	}

	get := func(field string) []string {
		vals, _ := comment.Get(field)
		return vals
	}
	if got := get("DATE"); len(got) != 1 || got[0] != "2020" {
		t.Errorf("DATE not preserved: %v", got)
	}
	if got := get("TITLE"); len(got) != 1 || got[0] != "new title" {
		t.Errorf("TITLE = %v, want [new title]", got)
	}
	if got := get("ALBUM"); len(got) != 1 || got[0] != "new album" {
		t.Errorf("ALBUM = %v, want [new album]", got)
	}
	if got := get("ARTIST"); len(got) != 1 || got[0] != "artistA" {
		t.Errorf("ARTIST = %v, want [artistA]", got)
	}
}

// TestUpdateMetaFlacAddsCoverWhenNoneExists confirms a single cover is added to
// a file that had none.
func TestUpdateMetaFlacAddsCoverWhenNoneExists(t *testing.T) {
	src := buildFLAC(t, map[string]string{"TITLE": "old"}, false)
	out := filepath.Join(t.TempDir(), "out.flac")

	params := &UpdateMetadataParams{
		Audio:       src,
		AudioExt:    ".flac",
		Meta:        fakeMeta{title: "t"},
		AlbumArt:    tinyPNG(t),
		AlbumArtExt: ".png",
	}
	if err := updateMetaFlac(context.Background(), out, params, zap.NewNop()); err != nil {
		t.Fatal(err)
	}

	res, err := flac.ParseFile(out)
	if err != nil {
		t.Fatal(err)
	}
	covers := 0
	for _, b := range res.Meta {
		if b.Type == flac.Picture {
			covers++
		}
	}
	if covers != 1 {
		t.Errorf("expected exactly 1 cover block, got %d", covers)
	}
}
