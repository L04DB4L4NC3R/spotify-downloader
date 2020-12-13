package pkg

type Repository interface {
	SaveMeta(songmeta *songMetaStruct, status string)
	SaveMetaArray(resource string, id string, songmeta []songMetaStruct, status string)
	FindMeta(resource string, id string) (*songMetaStruct, error)
	UpdateStatus(resource string, id string, status string)
	GetStatus(resource string, id string) (string, error)
}
