package core

import (
	"encoding/json"
	"strconv"
	"sync"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/scraper/proto"
	"github.com/rapito/go-spotify/spotify"
)

const (
	STATUS_META_FED     = "FED"
	STATUS_YT_FETCHED   = "FETCHED"
	STATUS_DWN_QUEUED   = "QUEUED"
	STATUS_DWN_COMPLETE = "COMPLETE"
	STATUS_DWN_FAILED   = "FAILED"
)

const (
	SPOT_TRACK_URL    = "https://open.spotify.com/track/"
	SPOT_ALBUM_URL    = "https://open.spotify.com/album/"
	SPOT_ARTIST_URL   = "https://open.spotify.com/artist/"
	SPOT_PLAYLIST_URL = "https://open.spotify.com/playlist/"

	RESOURCE_PLAYLIST = "playlists"
	RESOURCE_SONG     = "tracks"
	RESOURCE_ALBUM    = "albums"
)

const (
	PLAYLIST_BATCH_LIMIT = 1000
	PLAYLIST_BATCH_SIZE  = 100 // cannot be greater than this since spotify api playlist track limit is 100
)

type Service interface {
	// modules
	// takes albumn link and gives a link of song URLs
	FetchPlaylistMeta(resource string, id string) (songmetas []SongMeta, err error)
	// takes song URL and gives its metadata
	FetchSongMeta(id string) (*SongMeta, error)
	// Send a gRPC call to the ytber backend for further processing
	queueSongDownloadMessenger(_ *SongMeta, path *string) error
	queuePlaylistDownloadMessenger(resource string, id string, songmetas []SongMeta, path *string) error

	// core services
	SongDownload(id string, path *string) (*SongMeta, error)
	PlaylistDownload(resource string, id string, path *string) ([]SongMeta, error)
	PlaylistSync(url string, path *string) error

	// status tracking
	CheckSongStatus(id string) (string, error)
	CheckPlaylistStatus(resource string, id string) (string, error)
}

type service struct {
	redis               Repository
	spotify             *spotify.Spotify
	feedMetaTransporter pb.Service
}

func NewService(r Repository, s *spotify.Spotify, feedMetaTransporter pb.Service) Service {
	return &service{
		redis:               r,
		spotify:             s,
		feedMetaTransporter: feedMetaTransporter,
	}
}

// modules
// takes albumn link and gives a link of song URLs
func (s *service) FetchPlaylistMeta(resource string, id string) (songmetas []SongMeta, err error) {

	var (
		wg    sync.WaitGroup
		errs  []error
		query string
	)

	// if resource is album then no pagination (album API doesn't support pagination)
	if resource == RESOURCE_ALBUM {
		query = resource + "/%s"
		result, errs := s.spotify.Get(query, nil, id)
		if len(errs) != 0 {
			return nil, errs[0]
		}
		//log.Println(string(result))
		obj := SpotifyAlbumUnmarshalStruct{}
		err = json.Unmarshal(result, &obj)
		if err != nil {
			return nil, err
		}
		for _, val := range obj.Tracks.Items {
			songmetas = append(songmetas, SongMeta{
				Title:      val.Name,
				Url:        SPOT_TRACK_URL + val.ID,
				ArtistLink: SPOT_ARTIST_URL + val.Artists[0].ID,
				ArtistName: val.Artists[0].Name,
				AlbumName:  obj.Name,
				AlbumUrl:   SPOT_ALBUM_URL + id,
				Date:       obj.Date,
				Duration:   &val.DurationMs,
				Track:      &val.Track,
				SongID:     val.ID,
				Thumbnail:  obj.Images[0].Url,
			})
		}
		return songmetas, nil
	}

	// if 1000 is the limit and spotify api limit is 100 then call 10 goroutines
	wg.Add(PLAYLIST_BATCH_LIMIT/PLAYLIST_BATCH_SIZE + 1)
	query = resource + "/%s/tracks?limit=%s&offset=%s"
	for offset := 0; offset <= PLAYLIST_BATCH_LIMIT; offset += 100 {

		// max songs that can be returned in a playlist is 100
		go func(offset string) {

			result, errarr := s.spotify.Get(query, nil, id, "100", offset)
			if len(errarr) != 0 {
				errs = append(errs, errarr...)
				wg.Done()
				return
			}

			obj := SpotifyPlaylistUnmarshalStruct{}
			err = json.Unmarshal(result, &obj)
			if err != nil {
				errs = append(errs, err)
				wg.Done()
				return
			}

			for _, val := range obj.Items {
				songmetas = append(songmetas, SongMeta{

					Title:      val.Track.Name,
					Url:        SPOT_TRACK_URL + val.Track.ID,
					ArtistLink: SPOT_ARTIST_URL + val.Track.Album.Artists[0].ID,
					ArtistName: val.Track.Album.Artists[0].Name,
					AlbumName:  val.Track.Album.Name,
					AlbumUrl:   SPOT_ALBUM_URL + val.Track.Album.ID,
					Date:       val.Track.Album.Date,
					Duration:   &val.Track.DurationMs,
					Track:      &val.Track.Track,
					SongID:     val.Track.ID,
					Thumbnail:  val.Track.Album.Images[0].Url,
				})
			}
			wg.Done()
		}(strconv.Itoa(offset))
	}

	wg.Wait()
	if len(errs) != 0 {
		return nil, errs[0]
	}
	return songmetas, nil
}

// takes song URL and gives its metadata
func (s *service) FetchSongMeta(id string) (*SongMeta, error) {
	result, errs := s.spotify.Get("tracks/%s", nil, id)
	if len(errs) != 0 {
		return nil, errs[0]
	}
	obj := SpotifyUnmarshalStruct{}

	err := json.Unmarshal(result, &obj)
	if err != nil {
		return nil, err
	}

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
func (s *service) queueSongDownloadMessenger(songmeta *SongMeta, path *string) error {
	data := s.feedMetaTransporter.NewSongMetaTransportStruct(
		songmeta.Url,
		songmeta.SongID,
		songmeta.Thumbnail,
		songmeta.Genre,
		songmeta.Date,
		songmeta.AlbumUrl,
		songmeta.AlbumName,
		songmeta.ArtistLink,
		songmeta.ArtistName,
		uint32(*songmeta.Duration),
		uint32(*songmeta.Track),
		songmeta.Title,
	)
	_, err := s.feedMetaTransporter.SendSongMeta(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) queuePlaylistDownloadMessenger(resource string, id string, songmetas []SongMeta, path *string) error {
	metas := s.feedMetaTransporter.NewPlaylistTransportStruct()
	for _, songmeta := range songmetas {

		data := s.feedMetaTransporter.NewSongMetaTransportStruct(
			songmeta.Url,
			songmeta.SongID,
			songmeta.Thumbnail,
			songmeta.Genre,
			songmeta.Date,
			songmeta.AlbumUrl,
			songmeta.AlbumName,
			songmeta.ArtistLink,
			songmeta.ArtistName,
			uint32(*songmeta.Duration),
			uint32(*songmeta.Track),
			songmeta.Title,
		)
		metas.Songs = append(metas.Songs, *data)
	}

	metas.Type = resource
	metas.ID = id

	_, err := s.feedMetaTransporter.SendPlaylistMeta(metas)
	if err != nil {
		return err
	}
	return nil

}

// core services
func (s *service) SongDownload(id string, path *string) (*SongMeta, error) {
	songmeta, err := s.FetchSongMeta(id)
	if err != nil {
		return songmeta, err
	}
	// if error occurs while saving to redis, it is handled by the global async error handler
	go s.redis.SaveMeta(songmeta, STATUS_META_FED)
	return songmeta, s.queueSongDownloadMessenger(songmeta, path)
}

func (s *service) PlaylistDownload(resource string, id string, path *string) ([]SongMeta, error) {
	songmetas, err := s.FetchPlaylistMeta(resource, id)
	if err != nil {
		return nil, err
	}
	go s.redis.SaveMetaArray(resource, id, songmetas, STATUS_META_FED)
	return songmetas, s.queuePlaylistDownloadMessenger(resource, id, songmetas, path)
}

func (s *service) PlaylistSync(url string, path *string) error {
	panic("not implemented") // TODO: Implement
}

func (s *service) CheckSongStatus(id string) (string, error) {
	return s.redis.GetStatus(RESOURCE_SONG, id)
}

func (s *service) CheckPlaylistStatus(resource string, id string) (string, error) {
	return s.redis.GetStatus(resource, id)
}
