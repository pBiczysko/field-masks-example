package server

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pBiczysko/field-masks-example/proto"
)

// Backend implements the protobuf interface
type Backend struct {
	mu         *sync.RWMutex
	categories []*pb.Category
}

// New initializes a new Backend struct.
func New() *Backend {
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

// AddCategory adds a category to the in-memory store.
func (b *Backend) AddCategory(ctx context.Context, in *pb.AddCategoryRequest) (*pb.AddCategoryResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	cat := &pb.Category{
		Id:         uuid.Must(uuid.NewV4()).String(),
		Name:       in.GetName(),
		Price:      in.GetPrice(),
		ExternalId: in.GetExternalId(),
	}
	b.categories = append(b.categories, cat)

	return &pb.AddCategoryResponse{
		Category: cat,
	}, nil
}

// ListCategories lists all categories in the store.
func (b *Backend) ListCategories(ctx context.Context, _ *pb.ListCategoryRequest) (*pb.ListCategoryResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []*pb.Category
	for _, cat := range b.categories {
		out = append(out, cat)
	}

	return &pb.ListCategoryResponse{
		Categories: out,
	}, nil
}

// UpdateCategory updates given category in the store.
func (b *Backend) UpdateCategory(ctx context.Context, in *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	id := in.GetId()
	var out *pb.Category
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, cat := range b.categories {
		if cat.GetId() == id {
			out = cat
		}
	}
	if out == nil {
		return nil, status.Errorf(codes.NotFound, "category %q not found", in.GetId())
	}

	out.Name = in.GetName()
	out.Price = in.GetPrice()
	out.ExternalId = in.GetExternalId()

	return &pb.UpdateCategoryResponse{
		Category: out,
	}, nil
}
