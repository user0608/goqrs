package services

import (
	"context"
	"encoding/json"
	"fmt"
	"goqrs/database"
	"goqrs/models"
	"goqrs/repositories"
	"goqrs/security"
	"goqrs/xstorage"
	"io"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/ksaucedo002/answer/errores"
	"github.com/ksaucedo002/kcheck"
	"gorm.io/gorm"
)

const max_num_for_group = 200

type CollectionService interface {
	GetAll(ctx context.Context, username string) ([]models.Collection, error)
	GetByID(ctx context.Context, username string, id string) (*models.Collection, error)
	Delete(ctx context.Context, collectionID string) error
	GetTags(ctx context.Context, collectionID string) ([]models.Tag, error)
	AddTag(ctx context.Context, tag *models.Tag) error
	RemoveTag(ctx context.Context, tagid string) error

	Create(ctx context.Context, c *models.Collection) error
	Update(ctx context.Context, c *models.Collection, username string) error
	SaveTemplate(ctx context.Context, collectionid string, stt *models.TemlateDetails, img io.Reader) (templateUuid string, err error)
}

type collection struct {
	repository repositories.CollectionRepository
	xdir       xstorage.StorageService
}

func NewCollectionService(r repositories.CollectionRepository, xstrdir xstorage.StorageService) CollectionService {
	return &collection{repository: r, xdir: xstrdir}
}
func (s *collection) GetAll(ctx context.Context, username string) ([]models.Collection, error) {
	return s.repository.FindAll(database.Conn(ctx), username)
}
func (s *collection) GetByID(ctx context.Context, username string, id string) (*models.Collection, error) {
	if username == "" || id == "" {
		return nil, errores.NewBadRequestf(nil, errores.ErrRecordNotFaund)
	}
	return s.repository.FindByID(database.Conn(ctx), username, id)
}
func (s *collection) Delete(ctx context.Context, collectionID string) error {
	chk := kcheck.Atom{Name: "Collection ID", Value: collectionID}
	if err := kcheck.ValidTarget("uuid", chk); err != nil {
		return errores.NewBadRequestf(nil, err.Error())
	}
	return s.repository.DeleteCollection(database.Conn(ctx), security.UserName(ctx), collectionID)
}
func (s *collection) GetTags(ctx context.Context, collectionID string) ([]models.Tag, error) {
	chk := kcheck.Atom{Name: "Collection ID", Value: collectionID}
	if err := kcheck.ValidTarget("uuid", chk); err != nil {
		return nil, errores.NewBadRequestf(nil, err.Error())
	}
	return s.repository.FindTags(database.Conn(ctx), security.UserName(ctx), collectionID)
}

func (s *collection) AddTag(ctx context.Context, tag *models.Tag) error {
	if err := kcheck.Valid(tag); err != nil {
		return errores.NewBadRequestf(nil, err.Error())
	}
	return s.repository.AddTag(database.Conn(ctx), tag)
}
func (s *collection) RemoveTag(ctx context.Context, tagid string) error {
	chk := kcheck.Atom{Name: "Tag id", Value: tagid}
	if err := kcheck.ValidTarget("uuid", chk); err != nil {
		return errores.NewBadRequestf(nil, err.Error())
	}
	return s.repository.RemoveTag(database.Conn(ctx), tagid)
}

func (s *collection) Create(ctx context.Context, c *models.Collection) error {
	if c.NumTickets <= 0 {
		return errores.NewBadRequestf(nil, "el numero de tickets no pude ser 0")
	}
	if c.NotBefore == nil {
		t := time.Now()
		c.NotBefore = &t
	}
	if err := kcheck.Valid(c); err != nil {
		return errores.NewBadRequestf(nil, err.Error())
	}
	return database.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.repository.CreateCollection(tx, c); err != nil {
			return err
		}
		var forinsert = c.NumTickets
		for i := 0; i < c.NumTickets; i += max_num_for_group {
			num := int(math.Min(max_num_for_group, float64(forinsert)))
			var rows = make([]models.Ticket, num)
			for r := 0; r < num; r++ {
				rows[r] = models.Ticket{
					ID:           uuid.NewString(),
					CollectionID: c.ID,
				}
			}
			if err := s.repository.InsertTickets(tx, rows); err != nil {
				return err
			}
			forinsert -= max_num_for_group
		}
		return nil
	})
}
func (s *collection) Update(ctx context.Context, c *models.Collection, username string) error {
	if username == "" {
		return errores.NewBadRequestf(nil, errores.ErrRecordNotFaund)
	}
	if err := kcheck.Valid(c); err != nil {
		return errores.NewBadRequestf(nil, err.Error())
	}
	return s.repository.UpdateCollection(database.Conn(ctx), c, username)
}

func (s *collection) SaveTemplate(ctx context.Context, id string, stt *models.TemlateDetails, img io.Reader) (string, error) {
	if id == "" {
		return "", errores.NewBadRequestf(nil, "collection id no encontrado")
	}
	if stt.ItemWidth == 0 {
		return "", errores.NewBadRequestf(nil, "item_width no puede ser 0")
	}
	if stt.QqSize == 0 {
		return "", errores.NewBadRequestf(nil, "qr_size no puede ser 0")
	}
	username := security.UserName(ctx)
	jd, err := json.Marshal(stt)
	if err != nil {
		return "", errores.NewInternalf(nil, "no se pudo procesar las configuraciones")
	}
	tmpluuid := uuid.NewString()
	return tmpluuid, database.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.repository.SaveTemplateDetails(tx, id, username, tmpluuid, string(jd)); err != nil {
			return err
		}
		if err := s.xdir.Save(fmt.Sprintf("%s/%s.jpg", username, tmpluuid), img); err != nil {
			return errores.NewInternalf(err, errores.ErrRecord)
		}
		return nil
	})
}
