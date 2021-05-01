package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	pkg "github.com/L04DB4L4NC3R/spotify-downloader/scraper/pkg/core"
	"github.com/urfave/cli/v2"
)

const (
	PING     = "/ping/"
	SONG     = "/song/%s/"
	PLAYLIST = "/playlist/%s/"
	ALBUM    = "/album/%s/"
	SHOW     = "/show/%s/"

	SONG_STATUS = "/status/song/%s/"
)

var (
	ErrPingFail        = errors.New("ping failed")
	ErrInvalidArgCount = errors.New("invalid number of arguments")
	ErrSongFailed      = errors.New("song failed")
)

type status struct {
	Message string `json:"message"`
	Status  string `json:"data"`
}

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
		Name:  "ping",
		Usage: "check to see whether system is setup",
		Action: func(*cli.Context) error {
			then := time.Now()
			resp, err := h.client.Get(h.endpoint + PING)
			if err != nil {
				fmt.Printf("ping failed with error: %v", err)
				return err
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("ping failed with status: %d\n", resp.StatusCode)
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
		Name:    "song",
		Aliases: []string{"s", "track", "t"},
		Usage:   "download song",
		Action: func(c *cli.Context) error {
			then := time.Now()
			if c.NArg() < 1 {
				return ErrInvalidArgCount
			}
			resp, err := h.client.Get(
				h.endpoint +
					fmt.Sprintf(SONG, c.Args().Get(0)),
			)
			if err != nil {
				fmt.Printf("song failed with error: %v", err)
				return err
			}
			if resp.StatusCode == http.StatusInternalServerError {
				fmt.Printf("song failed with status: %d\n", resp.StatusCode)
				return ErrSongFailed
			}
			var previousStatus string
			for {
				resp, err := h.client.Get(
					h.endpoint +
						fmt.Sprintf(SONG_STATUS, c.Args().Get(0)),
				)
				if err == nil {
					var data status
					body, _ := ioutil.ReadAll(resp.Body)
					json.Unmarshal(body, &data)
					if previousStatus != data.Status {
						fmt.Printf("status: %s\n", data.Status)
					}
					if data.Status == pkg.STATUS_DWN_FAILED || data.Status == pkg.STATUS_DWN_COMPLETE {
						break
					}
					previousStatus = data.Status
				}
				time.Sleep(time.Duration(1) * time.Second)
			}
			fmt.Printf("song succeeded\nEndpoint: %s\nTime:%.4f ms",
				h.endpoint,
				float64(time.Since(then).Microseconds())/10000,
			)
			return nil
		},
	}
}

func (h *handler) DownloadPlaylist() *cli.Command {
	return &cli.Command{
		Name:    "playlist",
		Aliases: []string{"p"},
		Usage:   "download playlist",
		Action: func(c *cli.Context) error {
			then := time.Now()
			if c.NArg() < 1 {
				return ErrInvalidArgCount
			}
			resp, err := h.client.Get(
				h.endpoint +
					fmt.Sprintf(PLAYLIST, c.Args().Get(0)),
			)
			if err != nil {
				fmt.Printf("playlist failed with error: %v", err)
				return err
			}
			if resp.StatusCode == http.StatusInternalServerError {
				fmt.Printf("playlist failed with status: %d\n", resp.StatusCode)
				return ErrSongFailed
			}
			fmt.Printf("playlist succeeded\nEndpoint: %s\nTime:%.4f ms",
				h.endpoint,
				float64(time.Since(then).Microseconds())/10000,
			)
			return nil
		},
	}
}

func (h *handler) DownloadAlbum() *cli.Command {
	return &cli.Command{
		Name:    "album",
		Aliases: []string{"a"},
		Usage:   "download album",
		Action: func(c *cli.Context) error {
			then := time.Now()
			if c.NArg() < 1 {
				return ErrInvalidArgCount
			}
			resp, err := h.client.Get(
				h.endpoint +
					fmt.Sprintf(ALBUM, c.Args().Get(0)),
			)
			if err != nil {
				fmt.Printf("album failed with error: %v", err)
				return err
			}
			if resp.StatusCode == http.StatusInternalServerError {
				fmt.Printf("album failed with status: %d\n", resp.StatusCode)
				return ErrSongFailed
			}
			fmt.Printf("album succeeded\nEndpoint: %s\nTime:%.4f ms",
				h.endpoint,
				float64(time.Since(then).Microseconds())/10000,
			)
			return nil
		},
	}
}

func (h *handler) DownloadShow() *cli.Command {
	return &cli.Command{
		Name:    "show",
		Aliases: []string{"podcast"},
		Usage:   "download show",
		Action: func(c *cli.Context) error {
			then := time.Now()
			if c.NArg() < 1 {
				return ErrInvalidArgCount
			}
			resp, err := h.client.Get(
				h.endpoint +
					fmt.Sprintf(SHOW, c.Args().Get(0)),
			)
			if err != nil {
				fmt.Printf("show failed with error: %v", err)
				return err
			}
			if resp.StatusCode == http.StatusInternalServerError {
				fmt.Printf("show failed with status: %d\n", resp.StatusCode)
				return ErrSongFailed
			}
			fmt.Printf("show succeeded\nEndpoint: %s\nTime:%.4f ms",
				h.endpoint,
				float64(time.Since(then).Microseconds())/10000,
			)
			return nil
		},
	}
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
