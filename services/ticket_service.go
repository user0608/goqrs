package services

import (
	"context"
	"goqrs/database"
	"goqrs/models"
	"goqrs/repositories"
	"goqrs/security"
	"time"

	"github.com/ksaucedo002/answer/errores"
	"github.com/ksaucedo002/kcheck"
	"gorm.io/gorm"
)

type TicketService interface {
	ConsultTicketByID(ctx context.Context, uuid string) (*models.Ticket, error)

	InvalidTicket(ctx context.Context, uuid string) error
	ClaimTicket(ctx context.Context, uuid string) error
}

type ticket struct {
	collection repositories.CollectionRepository
	tickets    repositories.TicketRepository
}

func NewTicketService(c repositories.CollectionRepository, t repositories.TicketRepository) TicketService {
	return &ticket{collection: c, tickets: t}
}

func (s *ticket) ConsultTicketByID(ctx context.Context, uuid string) (*models.Ticket, error) {
	chk := kcheck.Atom{Name: "UUID", Value: uuid}
	if err := kcheck.ValidTarget("uuid", chk); err != nil {
		return nil, errores.NewBadRequestf(nil, err.Error())
	}
	var username = security.UserName(ctx)
	tx := database.Conn(ctx)
	ticket, err := s.tickets.FindTicketByID(tx, uuid)
	if err != nil {
		return nil, err
	}
	collection, err := s.collection.FindByID(tx, username, ticket.CollectionID)
	if err != nil {
		return nil, err
	}
	tags, err := s.collection.FindTags(tx, username, ticket.CollectionID)
	if err != nil {
		return nil, err
	}
	collection.Tags = tags
	ticket.Collection = collection
	return ticket, nil
}

func (s *ticket) InvalidTicket(ctx context.Context, uuid string) error {
	chk := kcheck.Atom{Name: "UUID", Value: uuid}
	if err := kcheck.ValidTarget("uuid", chk); err != nil {
		return errores.NewBadRequestf(nil, err.Error())
	}
	username := security.UserName(ctx)
	return database.Transaction(ctx, func(tx *gorm.DB) error {
		ticket, err := s.tickets.FindTicketByID(tx, uuid)
		if err != nil {
			return err
		}
		if ticket.Reclaimed != nil {
			return errores.NewBadRequestf(nil, "el ticket ya fue reclamado")
		}
		collection, err := s.collection.FindByID(tx, username, ticket.CollectionID)
		if err != nil {
			return err
		}
		if collection.DeletedAt.Valid {
			return errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
		}
		if err := s.tickets.InvalidateTicket(tx, uuid); err != nil {
			return err
		}
		return nil
	})
}
func (s *ticket) ClaimTicket(ctx context.Context, uuid string) error {
	chk := kcheck.Atom{Name: "UUID", Value: uuid}
	if err := kcheck.ValidTarget("uuid", chk); err != nil {
		return errores.NewBadRequestf(nil, err.Error())
	}
	username := security.UserName(ctx)
	return database.Transaction(ctx, func(tx *gorm.DB) error {
		ticket, err := s.tickets.FindTicketByID(tx, uuid)
		if err != nil {
			return err
		}
		if ticket.Reclaimed != nil {
			return errores.NewBadRequestf(nil, "el ticket ya fue reclamado")
		}
		if ticket.Annulled != nil {
			return errores.NewBadRequestf(nil, "El ticket ha sido anulado")
		}

		collection, err := s.collection.FindByID(tx, username, ticket.CollectionID)
		if err != nil {
			return err
		}
		t := time.Now()
		if collection.DeletedAt.Valid {
			return errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
		}
		if collection.NotBefore != nil {
			if t.Before(*collection.NotBefore) {
				return errores.NewBadRequestf(nil, "El ticket aún no es válido")
			}
		}
		if collection.TimeOut != nil {
			if t.After(*collection.TimeOut) {
				return errores.NewBadRequestf(nil, "El ticket ha caducado")
			}
		}
		if err := s.tickets.ClaimTicket(tx, uuid); err != nil {
			return err
		}
		return nil
	})
}
