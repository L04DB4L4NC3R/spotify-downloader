package core

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/PuerkitoBio/goquery"
)

const (

	META_FED
	YT_FETCHED
	DWN_QUEUED
	DWN_COMPLETE
	ACK
)

const (
	SONG_URL    = "http://localhost:3000/song/6c6W25YoDGjTq3qSPOga5t/"
}

type Service interface {
	// takes song URL and gives its metadata
	scrapeSongMeta(id string) (*SongMeta, error)
	// Send a gRPC call to the ytber backend for further processing
	queueSongDownloadMessenger(_ *SongMeta, path *string) error
	// core services
	SongDownload(id string, path *string) (*SongMeta, error)
	

}

type service struct {
	repo Repository
}

 NewService(r Repository) Service {
	return &service{
		repo: r,
	}
}

// takes song URL and gives its metadata
func (s *service) scrapeSongMeta(id string) (*SongMeta, error) {
	url := SONG_URL
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
func (s *service) queueSongDownloadMessenger(_ *SongMeta, path *string) error {
	// TODO: Send a gRPC call and fire forget.
	// panic("not implemented") // TODO: Implement
	return nil
}

// core services
func (s *service) SongDownload(id string, path *string) (*SongMeta, error) {
	songmeta, err := s.scrapeSongMeta(id)
	if err != nil {
		return songmeta, err
	}
	return songmeta, s.queueSongDownloadMessenger(songmeta, path)
}

