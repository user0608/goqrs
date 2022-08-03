package repositories

import (
	"goqrs/models"

	"github.com/ksaucedo002/answer/errores"
	"gorm.io/gorm"
)

type TicketRepository interface {
	FindTickets(tx *gorm.DB, collectionID, username string) ([]models.Ticket, error)
}
type ticket struct {
}

func NewTicketRepository() TicketRepository {
	return &ticket{}
}

func (*ticket) FindTickets(tx *gorm.DB, id, username string) ([]models.Ticket, error) {
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
