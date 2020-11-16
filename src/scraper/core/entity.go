package core

type SongMeta struct {
	Title      string `json:"title"`
	ArtistLink string `json:"artist_link"`
	ArtistPage string `json:"artist_page"`
	ArtistName string `json:"artist_name"`
	Url        string `json:"url"`
	YtUrl      string `json:"yt_url"`
	SongID     string `json:"song_id"`
}
