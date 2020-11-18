package core

type Repository interface {
	SaveMeta(songmeta *SongMeta, status string)
	FindMeta(resource string, id string) (*SongMeta, error)
	UpdateStatus(resource string, id string, status string)
	GetStatus(resource string, id string) (string, error)
}
