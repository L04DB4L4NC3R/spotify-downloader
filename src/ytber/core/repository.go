package core

type Repository interface {
	SaveMeta(songmeta SongMeta) error
	FindMeta(id string) (SongMeta, error)
}
