package main

import (
	"fmt"
	"log"
	"os"

	"net/http"

	handler "github.com/L04DB4L4NC3R/spotify-downloader/scraper/cli/handlers"
	"github.com/urfave/cli/v2"
)

func main() {
	client := &http.Client{}
	endpoint := "http://localhost:3000"
	var commands = handler.NewHandler(client, endpoint)
	app := &cli.App{
		Name:  "sdl",
		Usage: "sdl [action] [resource] [id]",
		Action: func(c *cli.Context) error {
			fmt.Println("spotify-downloader")
			return nil
		},
		Commands: []*cli.Command{
			commands.Health(),
			commands.DownloadSong(),
			commands.DownloadPlaylist(),
			commands.DownloadAlbum(),
			commands.DownloadShow(),
			commands.SyncPlaylist(),
			commands.PausePlaylistDownload(),
			commands.ResumePlaylistDownload(),
			commands.ViewSongProgress(),
			commands.ViewPlaylistProgress(),
			commands.ViewAlbumProgress(),
			commands.ViewShowProgress(),
			commands.ViewPlaylistMeta(),
			commands.ViewAlbumMeta(),
			commands.ViewSongMeta(),
			commands.ViewShowMeta(),
			commands.PlayPauseSong(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
