package ffmpeg

import (
	"context"
	"mime"
	"os"
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
)

func updateMetaFlac(_ context.Context, outPath string, m *UpdateMetadataParams, logger *zap.Logger) error {
	f, err := flac.ParseFile(m.Audio)
	if err != nil {
		return err
	}

	// generate comment block
	comment := flacvorbis.MetaDataBlockVorbisComment{Vendor: "unlock-music.dev"}

	// add metadata
	title := m.Meta.GetTitle()
	if title != "" {
		if err := comment.Add(flacvorbis.FIELD_TITLE, title); err != nil {
			logger.Warn("add flac title failed", zap.Error(err))
		}
	}

	album := m.Meta.GetAlbum()
	if album != "" {
		if err := comment.Add(flacvorbis.FIELD_ALBUM, album); err != nil {
			logger.Warn("add flac album failed", zap.Error(err))
		}
	}

	artists := m.Meta.GetArtists()
	for _, artist := range artists {
		if err := comment.Add(flacvorbis.FIELD_ARTIST, artist); err != nil {
			logger.Warn("add flac artist failed", zap.Error(err))
		}
	}

	existCommentIdx := slices.IndexFunc(f.Meta, func(b *flac.MetaDataBlock) bool {
		return b.Type == flac.VorbisComment
	})
	if existCommentIdx >= 0 { // copy existing comment fields
		exist, err := flacvorbis.ParseFromMetaDataBlock(*f.Meta[existCommentIdx])
		if err != nil {
			logger.Warn("parse existing flac comment failed", zap.Error(err))
		} else {
			for _, s := range exist.Comments {
				if strings.HasPrefix(s, flacvorbis.FIELD_TITLE+"=") && title != "" ||
					strings.HasPrefix(s, flacvorbis.FIELD_ALBUM+"=") && album != "" ||
					strings.HasPrefix(s, flacvorbis.FIELD_ARTIST+"=") && len(artists) != 0 {
					continue
				}
				comment.Comments = append(comment.Comments, s)
			}
		}
	}

	// add / replace flac comment
	cmtBlock := comment.Marshal()
	if existCommentIdx < 0 {
		f.Meta = append(f.Meta, &cmtBlock)
	} else {
		f.Meta[existCommentIdx] = &cmtBlock
	}

	if m.AlbumArt != nil {
		coverMime := mime.TypeByExtension(m.AlbumArtExt)
		logger.Debug("cover image mime detect", zap.String("mime", coverMime))
		cover, err := flacpicture.NewFromImageData(
			flacpicture.PictureTypeFrontCover,
			"Front cover",
			m.AlbumArt,
			coverMime,
		)

		if err != nil {
			logger.Warn("failed to create flac cover", zap.Error(err))
		} else {
			coverBlock := cover.Marshal()
			// add / replace flac cover
			coverIdx := slices.IndexFunc(f.Meta, func(b *flac.MetaDataBlock) bool {
				return b.Type == flac.Picture
			})
			if coverIdx < 0 {
				f.Meta = append(f.Meta, &coverBlock)
			} else {
				f.Meta[coverIdx] = &coverBlock
			}
		}
	}

	// Save atomically: write to a sibling temp file then rename over the target,
	// so a failure mid-write can't leave a truncated or corrupt output file.
	tmpPath := outPath + ".tmp"
	if err := f.Save(tmpPath); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, outPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
