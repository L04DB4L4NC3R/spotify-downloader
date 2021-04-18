package pkg

import (
	"context"
	"os/exec"
	"time"

	pb "github.com/L04DB4L4NC3R/spotify-downloader/ytber/proto"
	log "github.com/sirupsen/logrus"
)

type cmdError struct {
	Err error
	Cmd string
}

// ffmpeg encoding with thumbnail is best done serially
// after all the songs are downloaded
func (s *service) ApplyThumbnails(ctx context.Context, thumbscmds []string, resources []*pb.SongMetaRequest) {
	var (
		nctx, cancel = context.WithTimeout(ctx, THUMBNAIL_TEARDOWN_TIMEOUT)
		n            = len(thumbscmds)
		errChan      = make(chan *cmdError, n)
		cmdChan      = make(chan string, n)
		failedCmds   []string
	)

	log.Infof("queuing %d thumbnail jobs", n)

	// initialize worker pool for running commands
	defer cancel()
	go s.initExecPool(nctx, n, cmdChan, errChan)

	for _, command := range thumbscmds {
		cmdChan <- command
	}

	for _, v := range resources {
		cmdErr := <-errChan
		if cmdErr != nil {
			log.WithFields(log.Fields{
				"error": cmdErr.Err.Error(),
			}).Info("error executing thumbnail application")
			failedCmds = append(failedCmds, cmdErr.Cmd)
			s.redis.UpdateStatus(RESOURCE_SONG, v.SongId, STATUS_THUMBNAIL_FAILED)
		} else {
			log.Infof("thumnail succeeded for %s", v.Title)
			s.redis.UpdateStatus(RESOURCE_SONG, v.SongId, STATUS_FINISHED)
		}
	}
	failedCount := len(failedCmds)
	log.Infof("%d thumbnails applied, %d failed", n-failedCount, failedCount)
	ctx.Done()
}

func (s *service) initExecPool(
	ctx context.Context,
	count int,
	cmdChan chan string, // context bound cmds
	errChan chan *cmdError,
) {
	if count > MAXPROCS {
		count = MAXPROCS
	}
	for i := 0; i < count; i++ {
		go func() {
			for command := range cmdChan {
				nctx, cancel := context.WithTimeout(context.Background(), THUMBNAIL_APPLICATION_TIMEOUT)
				defer cancel()
				cmd := exec.CommandContext(nctx, "sh", "-c", command)
				if err := cmd.Run(); err != nil {
					errChan <- &cmdError{
						Err: err,
						Cmd: command,
					}
				}
				errChan <- nil
			}
		}()
	}
	// wait for context to signal teardown
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second * time.Duration(5))
		}
	}
}
