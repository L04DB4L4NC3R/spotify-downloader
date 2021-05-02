package core

type Repository interface {
	SaveMeta(songmeta *SongMeta, status string)
	SaveMetaArray(resource string, id string, songmeta []SongMeta, status string)
	FetchMetaArray(resource string, id string) (*PlaylistPayload, error)
	FindMeta(resource string, id string) (*SongMeta, error)
	UpdateStatus(resource string, id string, status string)
	GetStatus(resource string, id string) (string, error)
	GetBulkStatus(resource string, id []string) ([]string, error)
}
