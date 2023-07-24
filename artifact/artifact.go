package artifact

import (
	"context"
	"io"
)

type StorageClient interface {
	UploadDir(string, string) error
	UploadObjects(string, ...string) error
	GetDir(string, string) error
	GetObjects(string, string, ...string) error
	NewReader(context.Context, string, string) (io.ReadCloser, error)
}
