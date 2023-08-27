package ports

import (
	"context"
	"goqrs/internal/dominio/models"
)

type TicketRepository interface {
	FindTickets(ctx context.Context, collectionID, username string) ([]models.Ticket, error)
	FindTicketByID(ctx context.Context, uuid string) (*models.Ticket, error)

	InvalidateTicket(ctx context.Context, uuid string) error
	ClaimTicket(ctx context.Context, uuid string) error
}
