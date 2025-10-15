package standard

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/neatflowcv/gk/internal/pkg/domain"
	"github.com/neatflowcv/gk/internal/pkg/filesystem"
)

var _ filesystem.Filesystem = (*Filesystem)(nil)

type Filesystem struct{}

func New() *Filesystem {
	return &Filesystem{}
}

func (f *Filesystem) WalkPath(root string) ([]*domain.File, error) {
	var ret []*domain.File

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			// 루트 밖이거나 상대 경로 산출 실패 시 에러
			return fmt.Errorf("rel: %w", err)
		}

		ret = append(ret, domain.NewFile(path, rel, entry.IsDir()))

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk path: %w", err)
	}

	return ret, nil
}
