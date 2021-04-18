package pkg

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
)

type asyncReturn struct {
	meta              *pb.SongMetaRequest
	query             string
	postprocessingcmd string
}

func (s *service) PlaylistDownload(ctx context.Context, req *pb.PlaylistMetaRequest) (*pb.PlaylistMetaResponse, error) {
	var count int = len(req.Songs)
	log.WithFields(log.Fields{
		"count": count,
	}).Info("received playlist download request")

	go func(count int) {
		PLAYLIST_BATCH_SIZE, _ := strconv.Atoi(os.Getenv("PLAYLIST_BATCH_SIZE"))
		if PLAYLIST_BATCH_SIZE == 0 {
			PLAYLIST_BATCH_SIZE = MAXPROCS
		}
		var (
			batchCount         int     = 1
			retryCount         int     = 0
			totalBatches       float64 = math.Ceil(float64(count) / float64(PLAYLIST_BATCH_SIZE))
			retries, resources []*pb.SongMetaRequest
			thumbscmds         []string
		)

		fetchChan := make(chan asyncReturn, PLAYLIST_BATCH_SIZE)
		defer close(fetchChan)

		dispatchTime := time.Now()

		for i := 0; i < count; i += PLAYLIST_BATCH_SIZE {
			var offset int = i + PLAYLIST_BATCH_SIZE
			if offset > count {
				offset = count
			}
			log.WithFields(log.Fields{
				"total_songs_count": count,
				"batch_number":      batchCount,
				"total_batches":     totalBatches,
				"batch_index":       i,
				"batch_offset":      offset,
			}).Info("playlist batch execution")
			results, retry, metas := s.offloadBatchToYoutubeDL(ctx, req.Songs[i:offset], fetchChan)
			thumbscmds = append(thumbscmds, results...)
			resources = append(resources, metas...)
			retries = append(retries, retry...)
			batchCount++
		}

		s.ApplyThumbnails(ctx, thumbscmds, resources)
		log.WithFields(log.Fields{
			"total_songs_count": count,
			"total_batches":     totalBatches,
			"batch_size":        PLAYLIST_BATCH_SIZE,
			"time_taken":        time.Since(dispatchTime).Minutes(),
		}).Info("Download Successful")
		// using a closure for retry
		func() {
			numRetries := len(retries)
			if numRetries != 0 && retryCount < MAX_RETRIES {
				retryCount++
				log.WithFields(log.Fields{
					"retry_backoff_time_seconds": RETRY_BACKOFF_TIME.Seconds(),
					"number_of_songs":            numRetries,
					"current_retry":              retryCount,
					"max_retry_count":            MAX_RETRIES,
				}).Info("retrying failed downloads with backoff")
				time.Sleep(RETRY_BACKOFF_TIME)
				_, _ = s.PlaylistDownload(
					ctx,
					&pb.PlaylistMetaRequest{
						Songs:      retries,
						ResourceId: req.ResourceId,
						Type:       req.Type,
					},
				)
			}
		}()
		s.redis.UpdateStatus(req.Type, req.ResourceId, STATUS_FINISHED)
	}(count)

	res := &pb.PlaylistMetaResponse{
		Success: true,
	}

	return res, nil
}

func (s *service) offloadBatchToYoutubeDL(ctx context.Context, slice []*pb.SongMetaRequest, fetchChan chan asyncReturn) (
	postprocessingcmds []string,
	retry []*pb.SongMetaRequest,
	resources []*pb.SongMetaRequest,
) {

	ctx, cancel := context.WithTimeout(context.Background(), BATCH_DOWNLOAD_TIMEOUT)
	defer cancel()

	// to see when all the songs of the current batch are downloaded
	var (
		batchSize = len(slice)
	)

	for _, v := range slice {
		go func(v *pb.SongMetaRequest) {
			query := fmt.Sprintf("%s - %s", v.Title, v.ArtistName)
			fetchChan <- s.offloadToYoutubeDL(ctx, "mp3", query, v)
		}(v)
	}

	count := 0
	for {
		select {
		case <-ctx.Done():
			return postprocessingcmds, retry, resources
		case result := <-fetchChan:
			if count == batchSize-1 {
				return postprocessingcmds, retry, resources
			}
			if result.postprocessingcmd != "" {
				postprocessingcmds = append(postprocessingcmds, result.postprocessingcmd)
				resources = append(resources, result.meta)
			} else {
				retry = append(retry, result.meta)
			}
			count++
		default:
			time.Sleep(SONG_DOWNLOAD_WAIT_DURATION)
		}
	}
}
