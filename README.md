# spotify-downloader
Download songs, playlists and albums, or sync in to your favourite tunes

## Core Features

- [X] Track download on a URL basis
- [X] Entire playlist download
- [X] Apply metadata on downloaded songs
- [X] Entire album download
- [X] Independently scale downloader and renderer
- [X] Configurable concurrency

## UI Features

- [ ] Web-UI for downloading
- [X] CLI

## Run using docker compose

* Edit the ./config/secret.env to expose relevent secrets to the containers

* Change script permissions

```sh
chmod +x ./scripts/docker-setup.sh
```

* Run

```sh
./scripts/docker-setup.sh
```

## Run Locally
Installation instructions in wiki

## Endpoints

| Function | Route |
|:--------:|:-----:|
| Check service health | /ping/ |
| Download Song | /song/{id}/ |
| Download Playlist | /playlist/{id}/ |
| Download Album | /album/{id}/ |
| View Song Metadata | /meta/song/{id}/ |
| View Playlist Metadata | /meta/playlist/{id}/ |
| View Album Metadata | /meta/album/{id}/ |
| Bulk View Resource Metadata | /metas/<resource>/{id}/ (where resource can be playlists, albums, shows)|
| Check Song Download Progress | /status/song/{id}/ |
| Check Bulk Song Download Progress | /status/songs/ (array of {"song_ids": []string} in POST) |
| Check Playlist Download Progress | /status/playlist/{id}/ |
| Check Album Download Progress | /status/album/{id}/ |

Note that the `{id}` mentioned here is the resource ID you get from spotify (from a track, album or playlist URL).

## Disclaimer
Read the [disclaimer](disclaimer.md) before using this software.

## Contibutions Welcome
Note that this repo is just the core backend of the service. **UI contributions are needed**. All contributions are welcome. Just fork and make a PR. If you are making a UI, create a new directory called `src/ui`.
