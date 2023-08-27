package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goqrs/database"
	"goqrs/internal/dominio/models"
	"goqrs/internal/dominio/ports"

	"goqrs/pdfqr"

	"goqrs/security"
	"goqrs/xstorage"
	"log"

	"github.com/google/uuid"
	"github.com/ksaucedo002/answer/errores"
)

var ErrProcess = errors.New("no se pudo actualizar el estado del proceso")

type DocumentService struct {
	ticketRepo     ports.TicketRepository
	collectionRepo ports.CollectionRepository
	documentStorer xstorage.StorageService
	templateStorer xstorage.StorageService
}

func NewDocumentService(
	ticketRepo ports.TicketRepository,
	collectionRepo ports.CollectionRepository,
	templateStorer xstorage.StorageService,
	documentStorer xstorage.StorageService,

) DocumentService {
	return DocumentService{
		ticketRepo:     ticketRepo,
		collectionRepo: collectionRepo,
		documentStorer: documentStorer,
		templateStorer: templateStorer,
	}
}
func (s *DocumentService) FindDetailsAndCodes(ctx context.Context, id string) (*models.TemlateDetails, []models.Ticket, error) {
	username := security.UserName(ctx)
	if id == "" {
		return nil, nil, errores.NewBadRequestf(nil, "collection id not found")
	}
	var tdetails models.TemlateDetails
	var tickets []models.Ticket
	err := database.WithTx(ctx, func(ctx context.Context) error {
		collection, err := s.collectionRepo.FindByID(ctx, username, id)
		if err != nil {
			return err
		}
		if collection.TemplateDetails == "" {
			return errores.NewBadRequestf(nil, "no se encontró la configuración del templarte")
		}
		if err := json.Unmarshal([]byte(collection.TemplateDetails), &tdetails); err != nil {
			return errores.NewBadRequestf(nil, "el formato de configuraciones del templarte es inválido")
		}
		tickets, err = s.ticketRepo.FindTickets(ctx, id, username)
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
func (s *DocumentService) GenerateDocument(ctx context.Context, collectionID string) (string, error) {
	username := security.UserName(ctx)
	collection, err := s.collectionRepo.FindByID(ctx, username, collectionID)
	if err != nil {
		return "", err
	}
	if collection.DocumentProcess == "processing" {
		return "", errores.NewBadRequestf(nil, "el documento está siendo procesado")
	}
	if collection.TemplateUuid == "" {
		return "", errores.NewBadRequestf(nil, "no se encontro el template")
	}
	if collection.DocumentUuid != "" {
		go func() {
			s.documentStorer.Delete(s.pathDoc(ctx, collection.DocumentUuid)) // si existe, removemos anterior pdf
		}()
	}
	details, tickets, err := s.FindDetailsAndCodes(ctx, collectionID)
	if err != nil {
		return "", err
	}
	codes := make([]string, len(tickets))
	for i, t := range tickets {
		codes[i] = t.ID
	}
	if len(codes) == 0 {
		return "", errores.NewBadRequestf(nil, "no se encontraron los codigos")
	}
	if err := s.collectionRepo.StartDocumentProcess(ctx, collectionID, username); err != nil {
		log.Println(err)
		return "", ErrProcess
	}
	docuuid, err := s.generateDocumentAndSave(ctx, collection.TemplateUuid, details, codes)
	if err != nil {
		if err := s.collectionRepo.EndDocumentProcessWithError(ctx, collectionID, username, err.Error()); err != nil {
			log.Println(err)
			return "", ErrProcess
		}
		return "", err
	}
	if err := s.collectionRepo.EndDocumentProcessSucess(ctx, collectionID, username, docuuid); err != nil {
		log.Println(err)
		return "", ErrProcess
	}
	return docuuid, nil
}
func (s *DocumentService) generateDocumentAndSave(
	ctx context.Context,
	tmpluuid string,
	dt *models.TemlateDetails,
	codes []string,
) (docuuid string, err error) {
	docuuid = uuid.NewString()
	template, err := s.templateStorer.Find(s.pathTemplate(ctx, tmpluuid))
	if err != err {
		log.Println("service.generateDocument:", err)
		return "", errors.New("no se pudo encontrar la imagen template")
	}
	doc, err := pdfqr.CreateDocument(template, codes, pdfqr.DocumentConfigs{
		ItemWith: dt.ItemWidth,
		QrSize:   dt.QqSize,
		QrXPos:   dt.QqXPos,
		QrYPos:   dt.QqYPos,
	})
	if err != nil {
		return docuuid, err
	}
	if err := s.documentStorer.Save(s.pathDoc(ctx, docuuid), doc); err != nil {
		return docuuid, err
	}
	return docuuid, nil
}
func (s *DocumentService) pathTemplate(ctx context.Context, uuid string) string {
	username := security.UserName(ctx)
	return fmt.Sprintf("%s/%s.jpg", username, uuid)
}

func (s *DocumentService) pathDoc(ctx context.Context, uuid string) string {
	username := security.UserName(ctx)
	return fmt.Sprintf("%s/%s.pdf", username, uuid)
}
