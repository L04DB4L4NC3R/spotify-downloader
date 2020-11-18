# spotify-downloader
Download songs, playlists and albums, or sync in to your favourite tunes

## Features

- [ ] Song download on a URL basis
- [ ] Entire album download
- [ ] Pause and resume download ability (through status queues)
- [ ] Album sync daemon (CRONed or need basis)
- [ ] Web-UI for bulk process status handling
- [ ] Streaming music playback

## How to run

* Configure secrets: Copy `config/scraper.env.sample` to `config/scraper.env` and fill the secrets

* Build and run
```sh
make
./bin/scraper
```

## Roadmap

* Single song download
	* Get song link
	* Get download location
	* Scrape metadata
	* Pass metadata as a gRPC call to the ytber
	* Search yt for the song
	* Use youtube-dl to download mp3
	* Apply metadata patches
	* Update status on redis for song every step along the way
* Album download
	* Create song links dump from album
	* For each song repeat the above process in batches of N
* Album Sync Daemon
	* TBD
* Web-UI
	* TBD
* Music Player
	* TBD
