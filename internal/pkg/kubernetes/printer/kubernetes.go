package printer

import (
	"context"
	"fmt"

	"github.com/neatflowcv/gk/internal/pkg/domain"
	"github.com/neatflowcv/gk/internal/pkg/kubernetes"
)

var _ kubernetes.Kubernetes = (*Kubernetes)(nil)

type Kubernetes struct{}

func NewKubernetes() *Kubernetes {
	return &Kubernetes{}
}

func (k *Kubernetes) ApplyFile(ctx context.Context, namespace string, file *domain.File) error {
	_, _ = fmt.Println("namespace", namespace, "file", file.Rel()) //nolint:forbidigo

	return nil
}

func (k *Kubernetes) CreateNamespace(ctx context.Context, namespace string) error {
	_, _ = fmt.Println("namespace", namespace) //nolint:forbidigo

	return nil
}
