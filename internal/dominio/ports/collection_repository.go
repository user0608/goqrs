package ports

import (
	"context"
	"goqrs/internal/dominio/models"
)

type CollectionRepository interface {
	FindAll(ctx context.Context, username string) ([]models.Collection, error)
	FindByID(ctx context.Context, username string, id string) (*models.Collection, error)
	CreateCollection(ctx context.Context, c *models.Collection) error
	DeleteCollection(ctx context.Context, username, id string) error

	StartDocumentProcess(ctx context.Context, id, username string) error
	EndDocumentProcessWithError(ctx context.Context, id, username, messageerror string) error
	EndDocumentProcessSucess(ctx context.Context, id, username, docuuid string) error

	UpdateCollection(ctx context.Context, c *models.Collection, username string) error

	FindTags(ctx context.Context, username, collectionid string) ([]models.Tag, error)
	AddTag(ctx context.Context, t *models.Tag) error
	RemoveTag(ctx context.Context, tagID string) error

	InsertTickets(ctx context.Context, tikets []models.Ticket) error
	SaveTemplateDetails(ctx context.Context, collectionID, username, tmpluuid, jsond string) error
}
