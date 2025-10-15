package filesystem

import "github.com/neatflowcv/gk/internal/pkg/domain"

type Filesystem interface {
	WalkPath(path string) ([]*domain.File, error)
}
