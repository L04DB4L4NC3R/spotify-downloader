package pb

import (
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

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

func Register() error {
	addr := os.Getenv("YTBER_GRPC_SERVER_ADDR")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	RegisterFeedMetaServer(srv, &service{})
	log.WithFields(log.Fields{
		"grpc_server": addr,
	}).Info("Started gRPC server")
	if err := srv.Serve(listener); err != nil {
		return err
	}
	return nil
}
