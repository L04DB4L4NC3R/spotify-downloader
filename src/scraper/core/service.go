package core

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/rapito/go-spotify/spotify"
	log "github.com/sirupsen/logrus"
)

const (
	STATUS_META_FED     = "FED"
	STATUS_YT_FETCHED   = "YT_FETCHED"
	STATUS_DWN_QUEUED   = "DWN_QUEUED"
	STATUS_DWN_COMPLETE = "DWN_COMPLETE"
	STATUS_ACK          = "ACK"
)

const (
	SPOT_TRACK_URL    = "https://open.spotify.com/track/"
	SPOT_ALBUM_URL    = "https://open.spotify.com/album/"
	SPOT_ARTIST_URL   = "https://open.spotify.com/artist/"
	SPOT_PLAYLIST_URL = "https://open.spotify.com/playlist/"
)

type Service interface {
	// modules
	// takes albumn link and gives a link of song URLs
	fetchSongsFromPlaylist(url string) (urls []string, err error)
	// takes song URL and gives its metadata
	scrapeSongMeta(id string) (*SongMeta, error)
	// Send a gRPC call to the ytber backend for further processing
	queueSongDownloadMessenger(_ *SongMeta, path *string) error

	// core services
	SongDownload(id string, path *string) (*SongMeta, error)
	PlaylistDownload(id string, path *string) ([]SongMeta, []error)
	PlaylistSync(url string, path *string) error
}

type service struct {
	redis   Repository
	spotify *spotify.Spotify
}

func NewService(r Repository, s *spotify.Spotify) Service {
	return &service{
		redis:   r,
		spotify: s,
	}
}

// modules
// takes albumn link and gives a link of song URLs
func (s *service) fetchSongsFromPlaylist(url string) (urls []string, err error) {
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

	var (
		name    string
		songUrl string
		songid  string
		songids []string
	)
	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		name, _ = s.Attr("property")
		if name == "music:song" {
			songUrl, _ = s.Attr("content")
			songid = strings.Split(songUrl, SPOT_TRACK_URL)[1]
			songids = append(songids, songid)
		}
	})

	return songids, nil
}

// takes song URL and gives its metadata
func (s *service) scrapeSongMeta(id string) (*SongMeta, error) {
	result, errs := s.spotify.Get("tracks/%s", nil, id)
	if len(errs) != 0 {
		return nil, errs[0]
	}
	obj := SpotifyUnmarshalStruct{}

	err := json.Unmarshal(result, &obj)
	if err != nil {
		log.Error("EROEKORJNJOELNAKNFSKKSAFNKNSFKLJ")
		return nil, err
	}
	log.Info(obj)

	songmeta := &SongMeta{

		Title:      obj.Name,
		Url:        SPOT_TRACK_URL + id,
		ArtistLink: SPOT_ARTIST_URL + obj.Album.Artists[0].ID,
		ArtistName: obj.Album.Artists[0].Name,
		AlbumName:  obj.Album.Name,
		AlbumUrl:   SPOT_ALBUM_URL + obj.Album.ID,
		Date:       obj.Album.Date,
		Duration:   &obj.DurationMs,
		Track:      &obj.Track,
		SongID:     id,
		Thumbnail:  obj.Album.Images[0].Url,
	}

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
	// if error occurs while saving to redis, it is handled by the global async error handler
	go s.redis.SaveMeta(songmeta, STATUS_META_FED)
	return songmeta, s.queueSongDownloadMessenger(songmeta, path)
}

func (s *service) PlaylistDownload(id string, path *string) ([]SongMeta, []error) {
	url := SPOT_PLAYLIST_URL + id
	songs, err := s.fetchSongsFromPlaylist(url)
	if err != nil {
		return nil, []error{err}
	}
	var (
		songmeta  *SongMeta
		wg        sync.WaitGroup
		errs      []error
		songmetas []SongMeta
	)
	wg.Add(len(songs))

	// TODO: Use a different song download function which accepts songIDs in channels and
	// propoages results in channels, it then passes the meta to the queue function in channels too
	for _, songid := range songs {
		go func(songid string) {
			songmeta, err = s.SongDownload(songid, path)
			if err != nil {
				errs = append(errs, err)
			}
			songmetas = append(songmetas, *songmeta)
			wg.Done()
		}(songid)
	}
	wg.Wait()
	go s.redis.SaveMetaArray("playlist", id, songmetas, STATUS_META_FED)
	return songmetas, errs
}

func (s *service) PlaylistSync(url string, path *string) error {
	panic("not implemented") // TODO: Implement
}
