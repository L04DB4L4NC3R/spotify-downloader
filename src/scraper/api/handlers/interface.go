package handler

import (
	"net/http"
)

type Handler interface {
	// downloader
	Health() http.Handler
	DownloadSong() http.Handler
	DownloadPlaylist() http.Handler
	SyncPlaylist() http.Handler
	ViewPlaylistMeta() http.Handler
	ViewSongMeta() http.Handler
	ViewProgressOfPlaylistDownload() http.Handler
	PausePlaylistDownload() http.Handler
	ResumePlaylistDownload() http.Handler

	// player
	PlayPauseSong() http.Handler
}
