package pb

import (
	context "context"
	"errors"
)

type service struct {
	feedMetaTransporter FeedMetaClient
}

type Service interface {
	SendSongMeta(*songMetaStruct) (ytlink string, err error)
	SendPlaylistMeta(playlistTransportStruct) (ytlinks []string, err error)
	NewSongMetaTransportStruct(url, songId, thumbnail, genre, date, albumUrl, albumName, artistLink, artistName string, duration, track uint32, title string) *songMetaStruct
	NewPlaylistTransportStruct() playlistTransportStruct
}

func NewService(mtc FeedMetaClient) Service {
	return &service{
		feedMetaTransporter: mtc,
	}
}

func (svc *service) NewSongMetaTransportStruct(url, songId, thumbnail, genre, date, albumUrl, albumName, artistLink, artistName string, duration, track uint32, title string) *songMetaStruct {
	return &songMetaStruct{
		Url:        url,
		SongId:     songId,
		Thumbnail:  thumbnail,
		Genre:      genre,
		Date:       date,
		AlbumUrl:   albumUrl,
		AlbumName:  albumName,
		ArtistLink: artistLink,
		ArtistName: artistName,
		Duration:   duration,
		Track:      track,
		Title:      title,
	}
}

func (svc *service) NewPlaylistTransportStruct() playlistTransportStruct {
	return playlistTransportStruct{}
}
func (svc *service) SendSongMeta(meta *songMetaStruct) (ytlink string, err error) {

	req := &SongMetaRequest{
		Url:        meta.Url,
		SongId:     meta.SongId,
		Thumbnail:  meta.Thumbnail,
		Genre:      meta.Genre,
		Date:       meta.Date,
		AlbumUrl:   meta.AlbumUrl,
		AlbumName:  meta.AlbumName,
		ArtistLink: meta.ArtistLink,
		ArtistName: meta.ArtistName,
		Duration:   meta.Duration,
		Track:      meta.Track,
		Title:      meta.Title,
	}
	res, err := svc.feedMetaTransporter.SongDownload(context.Background(), req)
	if err != nil {
		return "", nil
	}
	if ok := res.Success; !ok {
		return "", errors.New(res.ErrMsg)
	}
	return res.YtUrl, nil
}

func (svc *service) SendPlaylistMeta(metas playlistTransportStruct) ([]string, error) {

	var (
		requests []*SongMetaRequest
	)
	for _, meta := range metas.Songs {
		req := &SongMetaRequest{
			Url:        meta.Url,
			SongId:     meta.SongId,
			Thumbnail:  meta.Thumbnail,
			Genre:      meta.Genre,
			Date:       meta.Date,
			AlbumUrl:   meta.AlbumUrl,
			AlbumName:  meta.AlbumName,
			ArtistLink: meta.ArtistLink,
			ArtistName: meta.ArtistName,
			Duration:   meta.Duration,
			Track:      meta.Track,
			Title:      meta.Title,
		}
		requests = append(requests, req)
	}
	playlistRq := &PlaylistMetaRequest{
		Songs:      requests,
		ResourceId: metas.ID,
		Type:       metas.Type,
	}
	res, err := svc.feedMetaTransporter.PlaylistDownload(context.Background(), playlistRq)
	if err != nil {
		return nil, err
	}
	if ok := res.Success; !ok {
		return nil, errors.New(res.ErrMsgs[0])
	}
	return res.YtUrls, nil
}
