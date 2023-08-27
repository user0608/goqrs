package repositories

import (
	"context"
	"goqrs/database"
	"goqrs/internal/dominio/models"
	"goqrs/internal/dominio/ports"

	"github.com/ksaucedo002/answer/errores"
	"gorm.io/gorm"
)

type login struct{}

func NewLoginRepository() ports.LoginRepository {
	return &login{}
}

func (r *login) IncrementPasswordAttempt(ctx context.Context, username string) error {
	tx := database.Conn(ctx)
	account := models.Account{Username: username}
	rs := tx.Model(&account).Update("password_attempt", gorm.Expr("password_attempt + ?", 1))
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	return nil
}
func (r *login) ResetPasswordAttempt(ctx context.Context, username string) error {
	tx := database.Conn(ctx)
	account := models.Account{Username: username}
	rs := tx.Model(&account).Update("password_attempt", 0)
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	return nil
}
func (r *login) FindAccount(ctx context.Context, username string) (*models.Account, error) {
	tx := database.Conn(ctx)
	var account models.Account
	rs := tx.Find(&account, "username=?", username)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return nil, errores.NewBadRequestf(nil, "usuario o contraseña invalidos")
	}
	return &account, nil
}
