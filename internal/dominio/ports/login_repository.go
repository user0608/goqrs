package ports

import (
	"context"
	"goqrs/internal/dominio/models"
)

type LoginRepository interface {
	IncrementPasswordAttempt(ctx context.Context, username string) error
	ResetPasswordAttempt(ctx context.Context, username string) error
	FindAccount(ctx context.Context, username string) (*models.Account, error)
}
