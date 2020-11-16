package core

const (
	META_SCRAPED = iota + 100
	META_FED
	YT_FETCHED
	DWN_QUEUED
	DWN_COMPLETE
	ACK
)

type Service interface {
	// modules
	// takes albumn link and gives a link of song URLs
	FetchSongsFromAlbum(url string) (urls []string, err error)
	// takes song URL and gives its metadata
	ScrapeSongMeta(url string) (*SongMeta, error)
	// Send a gRPC call to the ytber backend for further processing
	QueueSongDownloadMessenger(*SongMeta) error

	// core services
	SongDownload(url string, path string) error
	AlbumDownload(url string, path string) error
	AlbumSync(url string, path string) error
}
