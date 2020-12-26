package handler

import (
	"net/http"
)

type Handler interface {
	// downloader
	Health() http.Handler
	DownloadSong() http.Handler
	DownloadPlaylist() http.Handler
	DownloadAlbum() http.Handler
	SyncPlaylist() http.Handler

	// download state alterations
	PausePlaylistDownload() http.Handler
	ResumePlaylistDownload() http.Handler

	// download progress trackers
	ViewSongProgress() http.Handler

	// informational endpoints
	ViewPlaylistMeta() http.Handler
	ViewAlbumMeta() http.Handler
	ViewSongMeta() http.Handler

	// player
	PlayPauseSong() http.Handler
}
