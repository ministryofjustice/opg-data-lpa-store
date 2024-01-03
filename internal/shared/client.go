package shared

import (
	"context"
)

type Client interface {
	Put(ctx context.Context, data any) error
	Get(ctx context.Context, uid string) (Lpa, error)
}
