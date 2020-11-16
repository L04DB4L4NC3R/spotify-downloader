package handler

import (
	"net/http"
)

type Handler interface {
	// downloader
	Health() http.Handler
	DownloadSong() http.Handler
	DownloadAlbum() http.Handler
	SyncAlbum() http.Handler
	ViewAlbumMeta() http.Handler
	ViewSongMeta() http.Handler
	ViewProgressOfAlbumDownload() http.Handler
	PauseAlbumDownload() http.Handler
	ResumeAlbumDownload() http.Handler

	// player
	PlayPauseSong() http.Handler
}
