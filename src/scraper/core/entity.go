package core

type SongMeta struct {
	Title      string  `json:"title"`
	ArtistLink string  `json:"artist_link"`
	ArtistName string  `json:"artist_name"`
	AlbumName  string  `json:"album_name"`
	AlbumUrl   string  `json:"album_url"`
	Date       string  `json:"date"`
	Duration   *uint16 `json:"duration"`
	Bitrate    *uint8  `json:"bitrate"`
	Track      *uint8  `json:"track"`
	Genre      string  `json:"genre"`
	Url        string  `json:"url"`
	YtUrl      string  `json:"yt_url"`
	SongID     string  `json:"song_id"`
	Thumbnail  string  `json:"thumbnail"`
}
