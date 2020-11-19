package rpcservice

import (
	pb "github.com/L04DB4L4NC3R/spotify-downloader/scraper/rpc/proto"
)

type service struct {
	metaTransportClient *pb.FeedMetaClient
}

type Service interface {
}

func NewService(mtc *pb.FeedMetaClient) Service {
	return &service{
		metaTransportClient: mtc,
	}
}
