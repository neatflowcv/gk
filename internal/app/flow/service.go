package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neatflowcv/gk/internal/pkg/domain"
	"github.com/neatflowcv/gk/internal/pkg/filesystem"
	"github.com/neatflowcv/gk/internal/pkg/kubernetes"
)

type Service struct {
	filesystem filesystem.Filesystem
	kubernetes kubernetes.Kubernetes
}

func NewService(
	filesystem filesystem.Filesystem,
	kubernetes kubernetes.Kubernetes,
) *Service {
	return &Service{
		filesystem: filesystem,
		kubernetes: kubernetes,
	}
}

func (s *Service) ApplyPath(ctx context.Context, path string) error { //nolint:cyclop
	files, err := s.filesystem.WalkPath(path)
	if err != nil {
		return fmt.Errorf("파일 시스템 순회 실패: %w", err)
	}

	// namespace -> yaml files
	nsToFiles := make(map[string][]*domain.File)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		lower := strings.ToLower(file.Path())
		if !strings.HasSuffix(lower, ".yaml") && !strings.HasSuffix(lower, ".yml") {
			continue
		}

		rel, err := filepath.Rel(path, file.Path())
		if err != nil {
			// 루트 밖이거나 상대 경로 산출 실패 시 스킵
			continue
		}

		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) == 0 || parts[0] == "." || parts[0] == "" {
			// 루트 바로 아래에 파일이 있는 경우 네임스페이스를 정할 수 없음 → 스킵
			continue
		}

		ns := parts[0]
		nsToFiles[ns] = append(nsToFiles[ns], file)
	}

	var errs error

	for namespace, yamlFiles := range nsToFiles {
		err := s.kubernetes.CreateNamespace(ctx, namespace)
		if err != nil {
			errs = errors.Join(errs, err)
			_, _ = fmt.Fprintf(os.Stderr, "네임스페이스 보장 실패 [%s]: %v\n", namespace, err)

			continue
		}

		for _, yf := range yamlFiles {
			err := s.kubernetes.ApplyFile(ctx, namespace, yf)
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}
