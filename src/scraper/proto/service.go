package pb

type service struct {
	feedMetaTransporter *FeedMetaClient
}

type Service interface {
	SendSongMeta(map[string]interface{}) (ytlink string, err error)
	SendPlaylistMeta([]map[string]interface{}) (ytlinks []string, err []error)
}

func NewService(mtc *FeedMetaClient) Service {
	return &service{
		feedMetaTransporter: mtc,
	}
}

func (svc *service) SendSongMeta(meta map[string]interface{}) (ytlink string, err error) {
	/*
		req := pb.SongMetaRequest{
			Url:        meta["url"].(string),
			SongId:     meta["song_id"].(string),
			Thumbnail:  meta["thumbnail"].(string),
			Genre:      meta["genre"].(string),
			Date:       meta["date"].(string),
			AlbumUrl:   meta["album_url"].(string),
			AlbumName:  meta["album_name"].(string),
			ArtistLink: meta["album_link"].(string),
			ArtistName: meta["album_name"].(string),
			Duration:   meta["duration"].(uint32),
			Bitrate:    meta["bitrate"].(uint32),
			Track:      meta["track"].(uint32),
		}
		res, err := svc.feedMetaTransporter.SongDownload(context.Background(), req)
		if err != nil {
			return "", nil
		}
		return res.Result(), nil
	*/
	panic("NI")
}

func (svc *service) SendPlaylistMeta([]map[string]interface{}) (ytlinks []string, err []error) {
	panic("not implemented")
}
