package core

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type repo struct {
	Rdc  *redis.Client
	Cerr chan AsyncErrors
}

func NewRedisRepo(rdc *redis.Client, cerr chan AsyncErrors) Repository {
	return &repo{
		Rdc:  rdc,
		Cerr: cerr,
	}
}

// creates 2 fields in redis
// one for song metadata and the other for song status
func (r *repo) SaveMeta(songmeta *SongMeta, status string) {
	var (
		ctx context.Context = context.Background()
		err error
	)
	metakey := "song:meta:" + songmeta.SongID
	statuskey := "song:status:" + songmeta.SongID
	songmetabytes, _ := json.Marshal(songmeta)
	if err = r.Rdc.Set(ctx, metakey, string(songmetabytes), 0).Err(); err != nil {
		errobj := NewRepoError("Error saving meta", err, SRC_REDIS, metakey)
		r.Cerr <- errobj
		return
	}

	if err = r.Rdc.Set(ctx, statuskey, status, 0).Err(); err != nil {
		errobj := NewRepoError("Error saving meta", err, SRC_REDIS, metakey)
		r.Cerr <- errobj
		return
	}
	return
}

func (r *repo) FindMeta(resource string, id string) (*SongMeta, error) {
	metakey := resource + ":meta:" + id
	ctx := context.Background()
	val, err := r.Rdc.Get(ctx, metakey).Result()
	if err != nil {
		return nil, err
	}
	objectval := &SongMeta{}
	if err := json.Unmarshal([]byte(val), objectval); err != nil {
		return nil, err
	}
	return objectval, nil
}

func (r *repo) UpdateStatus(resource string, id string, status string) {
	var (
		ctx context.Context = context.Background()
		err error
	)
	statuskey := resource + ":status:" + id
	if err = r.Rdc.Set(ctx, statuskey, status, 0).Err(); err != nil {
		errobj := NewRepoError("Error updating status", err, SRC_REDIS, statuskey)
		r.Cerr <- errobj
		return
	}
	return
}

func (r *repo) GetStatus(resource string, id string) (string, error) {

	statuskey := resource + ":status:" + id
	ctx := context.Background()
	val, err := r.Rdc.Get(ctx, statuskey).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}
