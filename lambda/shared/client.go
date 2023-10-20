package shared

import (
	"context"
)

type Client interface {
	Put(ctx context.Context, data Case) error
	// Get(ctx context.Context, id string) (Case, error)
	// Patch(ctx context.Context, id string, data Update) (Case, error)
}
