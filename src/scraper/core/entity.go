package core

type SongMeta struct {
	Title      string `json:"title"`
	ArtistLink string `json:"artist_link"`
	ArtistPage string `json:"artist_page"`
	ArtistName string `json:"artist_name"`
	AlbumName  string `json:"album_name"`
	Date       string `json:"date"`
	Duration   string `json:"duration"`
	Start      string `json:"start"`
	Bitrate    uint8  `json:"bitrate"`
	Track      string `json:"track"`
	Genre      string `json:"genre"`
	Url        string `json:"url"`
	YtUrl      string `json:"yt_url"`
	SongID     string `json:"song_id"`
}
