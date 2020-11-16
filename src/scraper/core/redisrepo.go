package core

type repo struct {
}

func NewRedisRepo() Repository {
	return &repo{}
}

func (r *repo) SaveMeta(songmeta SongMeta) error {
	panic("not implemented") // TODO: Implement
}

func (r *repo) FindMeta(id string) (SongMeta, error) {
	panic("not implemented") // TODO: Implement
}
