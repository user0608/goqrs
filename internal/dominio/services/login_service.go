package services

import (
	"context"
	"errors"
	"goqrs/internal/dominio/models"
	"goqrs/internal/dominio/ports"

	"log"

	"github.com/ksaucedo002/answer/errores"
	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	r ports.LoginRepository
}

func NewLoginService(r ports.LoginRepository) LoginService {
	return LoginService{r: r}
}
func (s *LoginService) checkPassword(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errores.NewBadRequestf(nil, "usuario o contraseña invalidos")
		}
		return errores.NewBadRequestf(nil, "no se pudo decodificar la session")
	}
	return nil
}
func (s *LoginService) Login(ctx context.Context, username, password string) (*models.Account, error) {
	if username == "" || password == "" {
		return nil, errores.NewBadRequestf(nil, "usuario o contraseña no encontrados")
	}
	account, err := s.r.FindAccount(ctx, username)
	if err != nil {
		return nil, err
	}
	if account.PasswordAttempt > 8 {
		return nil, errores.NewBadRequestf(nil, "cuenta suspendida, por contraseña invalida")
	}
	if err := s.checkPassword(account.Password, password); err != nil {
		if err := s.r.IncrementPasswordAttempt(ctx, username); err != nil {
			log.Println(err)
		}
		return nil, err
	}
	if account.PasswordAttempt > 0 {
		if err := s.r.ResetPasswordAttempt(ctx, username); err != nil {
			log.Println(err)
		}
	}
	return account, nil
}
