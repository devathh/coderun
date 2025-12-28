package xcutrcontainer

import (
	"strings"

	customerrors "github.com/devathh/coderun/xcutr-service/pkg/errors"
)

type File struct {
	mime  string
	name  string
	bytes []byte
}

func NewFile(name, mime string, bytes []byte) (File, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return File{}, customerrors.ErrInvalidFilename
	}

	mime = strings.TrimSpace(mime)

	if len(bytes) < 1 {
		return File{}, customerrors.ErrEmptyFile
	}
	if len(bytes) > 1024*1024*60 {
		return File{}, customerrors.ErrTooLargeFile
	}

	return File{
		mime:  mime,
		name:  name,
		bytes: bytes,
	}, nil
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Mime() string {
	return f.mime
}

func (f *File) Bytes() []byte {
	return f.bytes
}
