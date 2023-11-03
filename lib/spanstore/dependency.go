package spanstore

import (
	"context"
	"errors"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
)

var (
	errNotImplemented = errors.New("not implemented")
)

type DependencyStore struct {
}

var _ dependencystore.Reader = (*DependencyStore)(nil)

func NewDependencyStore() *DependencyStore {
	return &DependencyStore{}
}

func (s *DependencyStore) GetDependencies(_ context.Context, _ time.Time, _ time.Duration) ([]model.DependencyLink, error) {
	return nil, errNotImplemented
}
