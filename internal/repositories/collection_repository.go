package repositories

import (
	"context"
	"fmt"
	"goqrs/database"
	"goqrs/internal/dominio/models"
	"goqrs/internal/dominio/ports"

	"github.com/ksaucedo002/answer/errores"
)

type collection struct{}

func NewCollectionRepository() ports.CollectionRepository {
	return &collection{}
}
func (r *collection) FindAll(ctx context.Context, username string) ([]models.Collection, error) {
	tx := database.Conn(ctx)
	var collections []models.Collection
	rs := tx.Where("account_username=?", username).Order("created_at desc").Find(&collections)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	return collections, nil
}

func (r *collection) FindByID(ctx context.Context, username string, id string) (*models.Collection, error) {
	tx := database.Conn(ctx)
	var collection models.Collection
	rs := tx.Where("account_username=? and id=?", username, id).Limit(1).Find(&collection)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return nil, errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return &collection, nil
}
func (r *collection) CreateCollection(ctx context.Context, c *models.Collection) error {
	tx := database.Conn(ctx)
	db := tx.Select("ID", "Name", "Description",
		"TimeOut", "NotBefore", "NumTickets",
		"AccountUsername")
	rs := db.Create(c)
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewInternalf(
			fmt.Errorf("error CreateCollection row=0 expected 1 row"),
			errores.ErrDatabaseInternal,
		)
	}
	return nil
}
func (r *collection) DeleteCollection(ctx context.Context, username, id string) error {
	tx := database.Conn(ctx)
	rs := tx.Where("id = ? and account_username = ?", id, username).Delete(&models.Collection{})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewBadRequestf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) UpdateCollection(ctx context.Context, c *models.Collection, username string) error {
	tx := database.Conn(ctx)
	db := tx.Select("Name", "Description", "TimeOut", "NotBefore")
	rs := db.Where("id = ? and account_username = ?", c.ID, username).Updates(c)
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewBadRequestf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) StartDocumentProcess(ctx context.Context, id, username string) error {
	tx := database.Conn(ctx)
	db := tx.Select("document_process").Where("id = ? and account_username = ?", id, username)
	rs := db.Updates(&models.Collection{
		DocumentProcess: "processing",
	})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewInternalf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) EndDocumentProcessWithError(ctx context.Context, id, username, messageerror string) error {
	tx := database.Conn(ctx)
	db := tx.Select("document_process", "process_result").Where("id = ? and account_username = ?", id, username)
	rs := db.Updates(&models.Collection{
		DocumentProcess: "error",
		ProcessResult:   messageerror,
	})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewInternalf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) EndDocumentProcessSucess(ctx context.Context, id, username, docuuid string) error {
	tx := database.Conn(ctx)
	db := tx.Select("document_process", "document_uuid").Where("id = ? and account_username = ?", id, username)
	rs := db.Updates(&models.Collection{
		DocumentProcess: "processed",
		DocumentUuid:    docuuid,
	})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewInternalf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) InsertTickets(ctx context.Context, tikets []models.Ticket) error {
	tx := database.Conn(ctx)
	rs := tx.Select("ID", "CollectionID").Create(&tikets)
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected != int64(len(tikets)) {
		return errores.NewInternalf(
			fmt.Errorf("error InsertTickets rows=%d expected %d row",
				rs.RowsAffected,
				len(tikets),
			),
			errores.ErrDatabaseInternal,
		)
	}
	return nil
}
func (r *collection) FindTags(ctx context.Context, username, collectionid string) ([]models.Tag, error) {
	tx := database.Conn(ctx)
	const qry = `select t.* from collection c inner join tag t
	on c.id = t.collection_id where c.id=? and c.account_username=?`
	rs := tx.Raw(qry, collectionid, username)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	var tags []models.Tag
	rs = rs.Scan(&tags)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	return tags, nil
}
func (r *collection) AddTag(ctx context.Context, t *models.Tag) error {
	tx := database.Conn(ctx)
	rs := tx.Create(t)
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) RemoveTag(ctx context.Context, tagID string) error {
	tx := database.Conn(ctx)
	rs := tx.Delete(&models.Tag{ID: tagID})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) SaveTemplateDetails(ctx context.Context, id, username, tmpluuid, jsond string) error {
	tx := database.Conn(ctx)
	rs := tx.Select("template_uuid", "template_details").
		Where("id = ? and account_username = ?", id, username).Updates(&models.Collection{
		TemplateUuid:    tmpluuid,
		TemplateDetails: jsond,
	})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
