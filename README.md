# spotify-downloader
Download songs, playlists and albums, or sync in to your favourite tunes

## Features

- [X] Song download on a URL basis
- [X] Entire playlist download
- [ ] Entire album download
- [ ] Pause and resume download ability
- [ ] Album sync daemon (CRON or need basis)
- [ ] Web-UI for bulk process status handling
- [ ] Streaming music playback

## How to run

* Configure secrets: Copy `config/local.env.sample` to `config/local.env` and fill the secrets

* Generate protocol buffer code (requires protoc & gRPC installation)

```sh
make build-proto
```

* Build and run
```sh
make run
```

* Kill

```sh
make kill
```

## Run using docker compose

* Configure secrets: Copy `config/docker.env.sample` to `config/docker.env` and fill the secrets

* Change script permissions

```sh
chmod +x ./scripts/docker-setup.sh
```

* Run

```sh
./scripts/docker-setup.sh
```

## Disclaimer
Read the [disclaimer](disclaimer.md) before using this software.
