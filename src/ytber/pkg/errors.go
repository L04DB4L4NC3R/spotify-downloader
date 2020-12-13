package pkg

type sources string

const (
	SRC_REDIS sources = "REDIS"
)

type AsyncErrors interface {
	Msg() string
	Err() error
	Src() sources
	Data() interface{}
}

type repoErrors struct {
	msg  string
	err  error
	src  sources
	data interface{}
}

func NewRepoError(msg string, err error, src sources, data interface{}) AsyncErrors {
	return &repoErrors{
		msg:  msg,
		err:  err,
		src:  src,
		data: data,
	}
}

func (ae *repoErrors) Msg() string {
	return ae.msg
}
func (ae *repoErrors) Err() error {
	return ae.err
}
func (ae *repoErrors) Src() sources {
	return ae.src
}
func (ae *repoErrors) Data() interface{} {
	return ae.data
}
