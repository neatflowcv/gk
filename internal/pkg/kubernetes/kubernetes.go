package kubernetes

import (
	"context"

	"github.com/neatflowcv/gk/internal/pkg/domain"
)

type Kubernetes interface {
	CreateNamespace(ctx context.Context, namespace string) error
	ApplyFile(ctx context.Context, namespace string, file *domain.File) error
}
