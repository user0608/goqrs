package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goqrs/database"
	"goqrs/models"
	"goqrs/pdfqr"
	"goqrs/repositories"
	"goqrs/security"
	"goqrs/xstorage"
	"log"

	"github.com/ksaucedo002/answer/errores"
	"gorm.io/gorm"
)

var ErrProcess = errors.New("no se pudo actualizar el estado del proceso")

type DocumentService interface {
	FindDetailsAndCodes(ctx context.Context, id string) (tdetails *models.TemlateDetails, tickets []models.Ticket, err error)
	GenerateDocument(ctx context.Context, id string, dt *models.TemlateDetails, codes []string) error
}
type docuement struct {
	ticketRepo     repositories.TicketRepository
	collectionRepo repositories.CollectionRepository
	documentStorer xstorage.StorageService
	templateStorer xstorage.StorageService
}

func NewDocumentService(
	ticketRepo repositories.TicketRepository,
	collectionRepo repositories.CollectionRepository,
	templateStorer xstorage.StorageService,
	documentStorer xstorage.StorageService,

) DocumentService {
	return &docuement{
		ticketRepo:     ticketRepo,
		collectionRepo: collectionRepo,
		documentStorer: documentStorer,
		templateStorer: templateStorer,
	}
}
func (s *docuement) FindDetailsAndCodes(ctx context.Context, id string) (*models.TemlateDetails, []models.Ticket, error) {
	username := security.UserName(ctx)
	if id == "" {
		return nil, nil, errores.NewBadRequestf(nil, "collection id not found")
	}
	var tdetails models.TemlateDetails
	var tickets []models.Ticket
	err := database.Transaction(ctx, func(tx *gorm.DB) error {
		collection, err := s.collectionRepo.FindByID(tx, username, id)
		if err != nil {
			return err
		}
		if collection.TemplateDetails == "" {
			return errores.NewBadRequestf(nil, "no se encontró la configuración del templarte")
		}
		if err := json.Unmarshal([]byte(collection.TemplateDetails), &tdetails); err != nil {
			return errores.NewBadRequestf(nil, "el formato de configuraciones del templarte es inválido")
		}
		tickets, err = s.ticketRepo.FindTickets(tx, id, username)
		if err != nil {
			return err
		}
		if len(tickets) == 0 {
			return errores.NewBadRequestf(nil, "no hay tickets para continuar con el proceso")
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return &tdetails, tickets, nil
}
func (s *docuement) GenerateDocument(ctx context.Context, id string, dt *models.TemlateDetails, codes []string) error {
	username := security.UserName(ctx)
	if err := s.collectionRepo.StartDocumentProcess(database.Conn(ctx), id, username); err != nil {
		log.Println(err)
		return ErrProcess
	}
	if err := s.generateDocument(ctx, id, dt, codes); err != nil {
		if err := s.collectionRepo.EndDocumentProcess(database.Conn(ctx), id, username, err.Error()); err != nil {
			log.Println(err)
			return ErrProcess
		}
		return err
	}
	if err := s.collectionRepo.EndDocumentProcess(database.Conn(ctx), id, username, ""); err != nil {
		log.Println(err)
		return ErrProcess
	}
	return nil
}
func (s *docuement) generateDocument(ctx context.Context, id string, dt *models.TemlateDetails, codes []string) error {
	username := security.UserName(ctx)
	template, err := s.templateStorer.Find(fmt.Sprintf("%s/%s.jpg", username, id))
	if err != err {
		log.Println("service.generateDocument:", err)
		return errors.New("no se pudo encontrar la imagen template")
	}
	doc, err := pdfqr.CreateDocument(template, codes, pdfqr.DocumentConfigs{
		ItemWith: dt.ItemWidth,
		QrSize:   dt.QqSize,
		QrXPos:   dt.QqXPos,
		QrYPos:   dt.QqYPos,
	})
	if err != nil {
		return err
	}
	if err := s.documentStorer.Save(fmt.Sprintf("%s/%s.pdf", username, id), doc); err != nil {
		return err
	}
	return nil
}
