package rpc

import (
	"os"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/scraper/rpc/proto"
	"google.golang.org/grpc"
)

func Register() (*pb.FeedMetaClient, error) {
	conn, err := grpc.Dial(os.Getenv("YTBER_GRPC_SERVER_ADDR"))
	if err != nil {
		return nil, err
	}
	client := pb.NewFeedMetaClient(conn)
	return &client, nil
}
