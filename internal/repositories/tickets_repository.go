package repositories

import (
	"context"
	"goqrs/database"
	"goqrs/internal/dominio/models"
	"goqrs/internal/dominio/ports"

	"time"

	"github.com/ksaucedo002/answer/errores"
)

type ticket struct {
}

func NewTicketRepository() ports.TicketRepository {
	return &ticket{}
}

func (*ticket) FindTickets(ctx context.Context, id, username string) ([]models.Ticket, error) {
	tx := database.Conn(ctx)
	const qry = `select t.* from collection c inner join ticket t on c.id = t.collection_id
	where c.id = ? and c.account_username = ?`
	rs := tx.Raw(qry, id, username)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	var tickets []models.Ticket
	rs = rs.Scan(&tickets)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return nil, errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return tickets, nil
}

func (*ticket) FindTicketByID(ctx context.Context, uuid string) (*models.Ticket, error) {
	tx := database.Conn(ctx)
	var ticket models.Ticket
	rs := tx.Where("id=?", uuid).Find(&ticket)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return nil, errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return &ticket, nil
}

func (*ticket) InvalidateTicket(ctx context.Context, uuid string) error {
	tx := database.Conn(ctx)
	rs := tx.Model(&models.Ticket{}).Where("id=?", uuid).Update("annulled", time.Now())
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewBadRequestf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}

func (*ticket) ClaimTicket(ctx context.Context, uuid string) error {
	tx := database.Conn(ctx)
	rs := tx.Model(&models.Ticket{}).Where("id=?", uuid).Update("reclaimed", time.Now())
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewBadRequestf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
