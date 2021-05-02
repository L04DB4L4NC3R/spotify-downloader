package core

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
)

type PlaylistPayload struct {
	SongMetas []SongMeta `json:"song_metas"`
}
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
	metakey := RESOURCE_SONG + ":meta:" + songmeta.SongID
	statuskey := RESOURCE_SONG + ":status:" + songmeta.SongID
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

func (r *repo) SaveMetaArray(resource string, id string, songmeta []SongMeta, status string) {
	var (
		ctx     context.Context = context.Background()
		err     error
		payload PlaylistPayload
	)
	metakey := resource + ":meta:" + id
	statuskey := resource + ":status:" + id
	payload.SongMetas = songmeta
	songmetabytes, _ := json.Marshal(payload)
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
	pipe := r.Rdc.Pipeline()
	for _, v := range songmeta {
		key := RESOURCE_SONG + ":status:" + v.SongID
		pipe.Set(ctx, key, STATUS_DWN_QUEUED, 0)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		errobj := NewRepoError("Error saving bulk meta", err, SRC_REDIS, metakey)
		r.Cerr <- errobj
		return
	}
	return
}

func (r *repo) FetchMetaArray(resource string, id string) (*PlaylistPayload, error) {
	var (
		metakey                 = resource + ":meta:" + id
		ctx     context.Context = context.Background()
		payload PlaylistPayload
	)
	data := r.Rdc.Get(ctx, metakey)
	if err := data.Err(); err != nil {
		log.Println(err.Error())
		return nil, err
	}
	b, err := data.Bytes()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	json.Unmarshal(b, &payload)
	return &payload, nil
}

func (r *repo) GetBulkStatus(resource string, id []string) ([]string, error) {
	var (
		status []string
		ops    []*redis.StringCmd
	)
	pipeliner := r.Rdc.Pipeline()
	defer pipeliner.Close()
	ctx := context.Background()
	for _, v := range id {
		key := resource + ":status:" + v
		ops = append(ops, pipeliner.Get(ctx, key))
	}
	if _, err := pipeliner.Exec(ctx); err != nil {
		return nil, err
	}
	for _, result := range ops {
		if r, e := result.Result(); e == nil {
			status = append(status, r)
		} else {
			status = append(status, STATUS_UNKNOWN)
		}
	}
	return status, nil
}
