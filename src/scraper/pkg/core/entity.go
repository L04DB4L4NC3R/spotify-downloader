package core

type SongMeta struct {
	Title      string `json:"title"`
	ArtistLink string `json:"artist_link"`
	ArtistName string `json:"artist_name"`
	AlbumName  string `json:"album_name"`
	AlbumUrl   string `json:"album_url"`
	Date       string `json:"date"`
	Duration   *int   `json:"duration"`
	Bitrate    *int   `json:"bitrate"`
	Track      *int   `json:"track"`
	Genre      string `json:"genre"`
	Url        string `json:"url"`
	YtUrl      string `json:"yt_url"`
	SongID     string `json:"song_id"`
	Thumbnail  string `json:"thumbnail"`
}

type SpotifyPlaylistUnmarshalStruct struct {
	Items []ItemSpotifyHelper `json:"items"`
}

type SpotifyAlbumUnmarshalStruct struct {
	Name   string                  `json:"name"`
	Images []ImageSpotifyHelper    `json:"images"`
	Date   string                  `json:"release_date"`
	Tracks SpotifyAlbumTrackStruct `json:"tracks"`
}

type ItemSpotifyHelper struct {
	Track SpotifyUnmarshalStruct `json:"track"`
}

type AlbumItemSpotifyHelper struct {
	AlbumSpotifyHelper
	Name       string `json:"name"`
	DurationMs int    `json:"duration_ms"`
	Track      int    `json:"track_number"`
	ID         string `json:"id"`
}

type SpotifyUnmarshalStruct struct {
	Name       string             `json:"name"`
	DurationMs int                `json:"duration_ms"`
	Track      int                `json:"track_number"`
	ID         string             `json:"id"`
	Album      AlbumSpotifyHelper `json:"album"`
}

type SpotifyAlbumTrackStruct struct {
	Items []AlbumItemSpotifyHelper `json:"items"`
}

type AlbumSpotifyHelper struct {
	ID          string                `json:"id"`
	Artists     []ArtistSpotifyHelper `json:"artists"`
	Images      []ImageSpotifyHelper  `json:"images"`
	Name        string                `json:"name"`
	Date        string                `json:"release_date"`
	TotalTracks int                   `json:"total_tracks"`
}
type ArtistSpotifyHelper struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
type ImageSpotifyHelper struct {
	Url string `json:"url"`
}
