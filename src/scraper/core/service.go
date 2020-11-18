package core

import (
	"encoding/json"

	"github.com/rapito/go-spotify/spotify"
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
	fetchSongsFromPlaylist(id string) (songmetas []SongMeta, err error)
	// takes song URL and gives its metadata
	scrapeSongMeta(id string) (*SongMeta, error)
	// Send a gRPC call to the ytber backend for further processing
	queueSongDownloadMessenger(_ *SongMeta, path *string) error

	// core services
	SongDownload(id string, path *string) (*SongMeta, error)
	PlaylistDownload(id string, path *string) ([]SongMeta, error)
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
func (s *service) fetchSongsFromPlaylist(id string) (songmetas []SongMeta, err error) {
	// max songs that can be returned in a playlist is 100
	// TODO: Create goroutines for paginated limit and skip returns
	result, errs := s.spotify.Get("playlists/%s/tracks?limit=%s&offset=%s", nil, id, "100", "0")
	if len(errs) != 0 {
		return nil, errs[0]
	}

	obj := SpotifyPlaylistUnmarshalStruct{}
	err = json.Unmarshal(result, &obj)
	if err != nil {
		return nil, err
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

func (s *service) PlaylistDownload(id string, path *string) ([]SongMeta, error) {
	songmetas, err := s.fetchSongsFromPlaylist(id)
	go s.redis.SaveMetaArray("playlist", id, songmetas, STATUS_META_FED)
	return songmetas, err
}

func (s *service) PlaylistSync(url string, path *string) error {
	panic("not implemented") // TODO: Implement
}
