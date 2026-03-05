package repository

import "context"

// Repository handles git operations for cloning and managing repos.
type Repository interface {
	Clone(ctx context.Context, url, dest, branch string) error
}
