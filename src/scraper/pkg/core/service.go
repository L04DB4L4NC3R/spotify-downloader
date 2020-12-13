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
)

const (
	PLAYLIST_BATCH_LIMIT = 1000
	PLAYLIST_BATCH_SIZE  = 100 // cannot be greater than this since spotify api playlist track limit is 100
)

type Service interface {
	// modules
	// takes albumn link and gives a link of song URLs
	fetchSongsFromPlaylist(id string) (songmetas []SongMeta, err error)
	// takes song URL and gives its metadata
	scrapeSongMeta(id string) (*SongMeta, error)
	// Send a gRPC call to the ytber backend for further processing
	queueSongDownloadMessenger(_ *SongMeta, path *string) error
	queuePlaylistDownloadMessenger(songmetas []SongMeta, path *string) error

	// core services
	SongDownload(id string, path *string) (*SongMeta, error)
	PlaylistDownload(id string, path *string) ([]SongMeta, error)
	PlaylistSync(url string, path *string) error
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
func (s *service) fetchSongsFromPlaylist(id string) (songmetas []SongMeta, err error) {

	var (
		wg   sync.WaitGroup
		errs []error
	)

	// if 1000 is the limit and spotify api limit is 100 then call 10 goroutines
	wg.Add(PLAYLIST_BATCH_LIMIT/PLAYLIST_BATCH_SIZE + 1)
	for offset := 0; offset <= PLAYLIST_BATCH_LIMIT; offset += 100 {

		// max songs that can be returned in a playlist is 100
		go func(offset string) {

			result, errarr := s.spotify.Get("playlists/%s/tracks?limit=%s&offset=%s", nil, id, "100", offset)
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
func (s *service) scrapeSongMeta(id string) (*SongMeta, error) {
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

func (s *service) queuePlaylistDownloadMessenger(songmetas []SongMeta, path *string) error {
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
		metas = append(metas, *data)
	}

	_, err := s.feedMetaTransporter.SendPlaylistMeta(metas)
	if err != nil {
		return err
	}
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

func (s *service) PlaylistDownload(id string, path *string) ([]SongMeta, error) {
	songmetas, err := s.fetchSongsFromPlaylist(id)
	if err != nil {
		return nil, err
	}
	// go s.redis.SaveMetaArray("playlist", id, songmetas, STATUS_META_FED)
	return songmetas, s.queuePlaylistDownloadMessenger(songmetas, path)
}

func (s *service) PlaylistSync(url string, path *string) error {
	panic("not implemented") // TODO: Implement
}
