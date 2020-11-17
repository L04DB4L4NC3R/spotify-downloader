package core

import (
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	META_SCRAPED = iota + 100
	META_FED
	YT_FETCHED
	DWN_QUEUED
	DWN_COMPLETE
	ACK
)

const (
	SPOT_TRACK_URL    = "https://open.spotify.com/track/"
	SPOT_PLAYLIST_URL = "https://open.spotify.com/playlist/"
)

type Service interface {
	// modules
	// takes albumn link and gives a link of song URLs
	fetchSongsFromAlbum(url string) (urls []string, err error)
	// takes song URL and gives its metadata
	scrapeSongMeta(id string) (*SongMeta, error)
	// Send a gRPC call to the ytber backend for further processing
	queueSongDownloadMessenger(*SongMeta) error

	// core services
	SongDownload(id string, path *string) (*SongMeta, error)
	AlbumDownload(url string, path string) error
	AlbumSync(url string, path string) error
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{
		repo: r,
	}
}

// modules
// takes albumn link and gives a link of song URLs
func (s *service) fetchSongsFromAlbum(url string) (urls []string, err error) {
	panic("not implemented") // TODO: Implement
}

// takes song URL and gives its metadata
func (s *service) scrapeSongMeta(id string) (*SongMeta, error) {
	url := SPOT_TRACK_URL + id
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	songmeta := &SongMeta{
		Url:    url,
		SongID: id,
	}

	var name string
	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		name, _ = s.Attr("property")
		if name == "og:title" {
			songmeta.Title, _ = s.Attr("content")
		} else if name == "music:musician" {
			songmeta.ArtistLink, _ = s.Attr("content")
		} else if name == "og:image" {
			songmeta.Thumbnail, _ = s.Attr("content")
		} else if name == "music:duration" {
			strduration, _ := s.Attr("content")
			duration, _ := strconv.ParseUint(strduration, 10, 16)
			typecastedDuration := uint16(duration)
			songmeta.Duration = &typecastedDuration
		} else if name == "music:album" {
			songmeta.AlbumUrl, _ = s.Attr("content")
		} else if name == "music:album:track" {
			strtrack, _ := s.Attr("content")
			track, _ := strconv.ParseUint(strtrack, 10, 8)
			typecastedTrack := uint8(track)
			songmeta.Track = &typecastedTrack
		} else if name == "music:release_date" {
			songmeta.Date, _ = s.Attr("content")
		} else if name == "twitter:audio:artist_name" {
			songmeta.ArtistName, _ = s.Attr("content")
		}

	})

	songmeta.AlbumName = doc.Find("div.media-bd a").Last().Text()
	return songmeta, err
}

// Send a gRPC call to the ytber backend for further processing
func (s *service) queueSongDownloadMessenger(_ *SongMeta) error {
	panic("not implemented") // TODO: Implement
}

// core services
func (s *service) SongDownload(id string, path *string) (*SongMeta, error) {
	// TODO: Send a gRPC call and fire forget.
	return s.scrapeSongMeta(id)
}

func (s *service) AlbumDownload(url string, path string) error {
	panic("not implemented") // TODO: Implement
}

func (s *service) AlbumSync(url string, path string) error {
	panic("not implemented") // TODO: Implement
}
