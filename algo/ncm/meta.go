package ncm

import (
	"strings"

	"go.uber.org/zap"

	"git.um-react.app/um/cli/algo/common"
)

type ncmMeta interface {
	common.AudioMeta

	// GetFormat return the audio format, e.g. mp3, flac
	GetFormat() string

	// GetAlbumImageURL return the album image url
	GetAlbumImageURL() string
}

type ncmMetaMusic struct {
	logger *zap.Logger

	Format        string `json:"format"`
	MusicName     string `json:"musicName"`
	Artist        any    `json:"artist"`
	Album         string `json:"album"`
	AlbumPicDocID any    `json:"albumPicDocId"`
	AlbumPic      string `json:"albumPic"`
	Flag          int    `json:"flag"`
	Bitrate       int    `json:"bitrate"`
	Duration      int    `json:"duration"`
	Alias         []any  `json:"alias"`
	TransNames    []any  `json:"transNames"`
}

func newNcmMetaMusic(logger *zap.Logger) *ncmMetaMusic {
	ncm := new(ncmMetaMusic)
	ncm.logger = logger.With(zap.String("module", "ncmMetaMusic"))
	return ncm
}

func (m *ncmMetaMusic) GetAlbumImageURL() string {
	return m.AlbumPic
}

func (m *ncmMetaMusic) GetArtists() []string {
	m.logger.Debug("ncm artists raw", zap.Any("artists", m.Artist))
	var artists []string
	switch v := m.Artist.(type) {

	// Simple format: "artistA"
	// Ref: https://git.unlock-music.dev/um/cli/issues/78
	case string:
		artists = []string{v}

	// Nested / mixed-type format, e.g. [["artistA", 12345], ["artistB", 67890]].
	// JSON unmarshalled into an `any` produces []any of []any, never [][]string.
	case []any:
		for _, item := range v {
			if innerSlice, ok := item.([]any); ok {
				if len(innerSlice) > 0 {
					// Assume the first element is the artist's name.
					if artistName, ok := innerSlice[0].(string); ok {
						artists = append(artists, artistName)
					}
				}
			}
		}

	default:
		// Log a warning if the artist type is unexpected and not handled.
		m.logger.Warn("unexpected artist type", zap.Any("artists", m.Artist))
	}

	return artists
}

func (m *ncmMetaMusic) GetTitle() string {
	return m.MusicName
}

func (m *ncmMetaMusic) GetAlbum() string {
	return m.Album
}

func (m *ncmMetaMusic) GetFormat() string {
	return m.Format
}

//goland:noinspection SpellCheckingInspection
type ncmMetaDJ struct {
	ProgramID          int          `json:"programId"`
	ProgramName        string       `json:"programName"`
	MainMusic          ncmMetaMusic `json:"mainMusic"`
	DjID               int          `json:"djId"`
	DjName             string       `json:"djName"`
	DjAvatarURL        string       `json:"djAvatarUrl"`
	CreateTime         int64        `json:"createTime"`
	Brand              string       `json:"brand"`
	Serial             int          `json:"serial"`
	ProgramDesc        string       `json:"programDesc"`
	ProgramFeeType     int          `json:"programFeeType"`
	ProgramBuyed       bool         `json:"programBuyed"`
	RadioID            int          `json:"radioId"`
	RadioName          string       `json:"radioName"`
	RadioCategory      string       `json:"radioCategory"`
	RadioCategoryID    int          `json:"radioCategoryId"`
	RadioDesc          string       `json:"radioDesc"`
	RadioFeeType       int          `json:"radioFeeType"`
	RadioFeeScope      int          `json:"radioFeeScope"`
	RadioBuyed         bool         `json:"radioBuyed"`
	RadioPrice         int          `json:"radioPrice"`
	RadioPurchaseCount int          `json:"radioPurchaseCount"`
}

func (m *ncmMetaDJ) GetArtists() []string {
	if m.DjName != "" {
		return []string{m.DjName}
	}
	return m.MainMusic.GetArtists()
}

func (m *ncmMetaDJ) GetTitle() string {
	if m.ProgramName != "" {
		return m.ProgramName
	}
	return m.MainMusic.GetTitle()
}

func (m *ncmMetaDJ) GetAlbum() string {
	if m.Brand != "" {
		return m.Brand
	}
	return m.MainMusic.GetAlbum()
}

func (m *ncmMetaDJ) GetFormat() string {
	return m.MainMusic.GetFormat()
}

func (m *ncmMetaDJ) GetAlbumImageURL() string {
	if strings.HasPrefix(m.MainMusic.GetAlbumImageURL(), "http") {
		return m.MainMusic.GetAlbumImageURL()
	}
	return m.DjAvatarURL
}
