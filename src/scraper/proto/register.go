package pb

import (
	"os"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

const (
	PLAYLIST uint = iota
	ALBUM
)

type playlistTransportStruct struct {
	Songs []songMetaStruct
	ID    string
	Type  string
}

type songMetaStruct struct {
	Url        string
	SongId     string
	Thumbnail  string
	Genre      string
	Date       string
	AlbumUrl   string
	AlbumName  string
	ArtistLink string
	ArtistName string
	Duration   uint32
	Bitrate    uint32
	Track      uint32
	Title      string
}

func Register() (*grpc.ClientConn, FeedMetaClient, error) {
	addr := os.Getenv("YTBER_GRPC_SERVER_ADDR")
	if addr == "" {
		addr = "localhost:4040"
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	client := NewFeedMetaClient(conn)
	log.WithFields(log.Fields{
		"grpc_server": addr,
	}).Info("Connected to the gRPC server")
	return conn, client, nil
}
