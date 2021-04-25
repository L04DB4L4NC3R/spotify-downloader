package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	PING = "/ping/"
)

var (
	ErrPingFail = errors.New("ping failed")
)

type handler struct {
	endpoint string
	client   *http.Client
}

func NewHandler(client *http.Client, endpoint string) Handler {
	return &handler{client: client, endpoint: endpoint}
}

// downloader
func (h *handler) Health() *cli.Command {
	return &cli.Command{
		Name:    "ping",
		Aliases: []string{"p"},
		Usage:   "check to see whether system is setup",
		Action: func(*cli.Context) error {
			then := time.Now()
			resp, err := h.client.Get(h.endpoint + PING)
			if err != nil {
				fmt.Printf("ping failed with error: %v", err)
				return err
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("ping failed with status: %d", resp.StatusCode)
				return ErrPingFail
			}
			fmt.Printf("ping succeeded\nEndpoint: %s\nTime:%.4f ms",
				h.endpoint,
				float64(time.Since(then).Microseconds())/10000,
			)
			return nil
		},
	}
}

func (h *handler) DownloadSong() *cli.Command {
	return &cli.Command{
		Name:        "",
		Aliases:     nil,
		Usage:       "",
		UsageText:   "",
		Description: "DownloadSong",
		Category:    "",
		Action:      func(*cli.Context) error { return nil },
	}
}

func (h *handler) DownloadPlaylist() *cli.Command {
	return &cli.Command{}
}

func (h *handler) DownloadAlbum() *cli.Command {
	return &cli.Command{}
}

func (h *handler) DownloadShow() *cli.Command {
	return &cli.Command{}
}

func (h *handler) SyncPlaylist() *cli.Command {
	return &cli.Command{}
}

// download state alterations
func (h *handler) PausePlaylistDownload() *cli.Command {
	return &cli.Command{}
}

func (h *handler) ResumePlaylistDownload() *cli.Command {
	return &cli.Command{}
}

// download progress trackers
func (h *handler) ViewSongProgress() *cli.Command {
	return &cli.Command{}
}

func (h *handler) ViewPlaylistProgress() *cli.Command {
	return &cli.Command{}
}

func (h *handler) ViewAlbumProgress() *cli.Command {
	return &cli.Command{}
}

func (h *handler) ViewShowProgress() *cli.Command {
	return &cli.Command{}
}

// informational endpoints
func (h *handler) ViewPlaylistMeta() *cli.Command {
	return &cli.Command{}
}

func (h *handler) ViewAlbumMeta() *cli.Command {
	return &cli.Command{}
}

func (h *handler) ViewSongMeta() *cli.Command {
	return &cli.Command{}
}

func (h *handler) ViewShowMeta() *cli.Command {
	return &cli.Command{}
}

// player
func (h *handler) PlayPauseSong() *cli.Command {
	return &cli.Command{}
}
