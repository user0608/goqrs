package repositories

import (
	"fmt"
	"goqrs/models"

	"github.com/ksaucedo002/answer/errores"
	"gorm.io/gorm"
)

type CollectionRepository interface {
	FindAll(tx *gorm.DB, username string) ([]models.Collection, error)
	FindByID(tx *gorm.DB, username string, id string) (*models.Collection, error)
	CreateCollection(tx *gorm.DB, c *models.Collection) error
	DeleteCollection(tx *gorm.DB, username, id string) error

	StartDocumentProcess(tx *gorm.DB, id, username string) error
	EndDocumentProcessWithError(tx *gorm.DB, id, username, messageerror string) error
	EndDocumentProcessSucess(tx *gorm.DB, id, username, docuuid string) error

	UpdateCollection(tx *gorm.DB, c *models.Collection, username string) error

	FindTags(tx *gorm.DB, username, collectionid string) ([]models.Tag, error)
	AddTag(tx *gorm.DB, t *models.Tag) error
	RemoveTag(tx *gorm.DB, tagID string) error

	InsertTickets(tx *gorm.DB, tikets []models.Ticket) error
	SaveTemplateDetails(tx *gorm.DB, collectionID, username, tmpluuid, jsond string) error
}

type collection struct {
}

func NewCollectionRepository() CollectionRepository {
	return &collection{}
}
func (r *collection) FindAll(tx *gorm.DB, username string) ([]models.Collection, error) {
	var collections []models.Collection
	rs := tx.Where("account_username=?", username).Order("created_at desc").Find(&collections)
	if rs.Error != nil {
		return nil, errores.NewInternalDBf(rs.Error)
	}
	return collections, nil
}

func (r *collection) FindByID(tx *gorm.DB, username string, id string) (*models.Collection, error) {
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
func (r *collection) CreateCollection(tx *gorm.DB, c *models.Collection) error {
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
func (r *collection) DeleteCollection(tx *gorm.DB, username, id string) error {
	rs := tx.Where("id = ? and account_username = ?", id, username).Delete(&models.Collection{})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewBadRequestf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) UpdateCollection(tx *gorm.DB, c *models.Collection, username string) error {
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
func (r *collection) StartDocumentProcess(tx *gorm.DB, id, username string) error {
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
func (r *collection) EndDocumentProcessWithError(tx *gorm.DB, id, username, messageerror string) error {
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
func (r *collection) EndDocumentProcessSucess(tx *gorm.DB, id, username, docuuid string) error {
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
func (r *collection) InsertTickets(tx *gorm.DB, tikets []models.Ticket) error {
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
func (r *collection) FindTags(tx *gorm.DB, username, collectionid string) ([]models.Tag, error) {
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
func (r *collection) AddTag(tx *gorm.DB, t *models.Tag) error {
	rs := tx.Create(t)
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) RemoveTag(tx *gorm.DB, tagID string) error {
	rs := tx.Delete(&models.Tag{ID: tagID})
	if rs.Error != nil {
		return errores.NewInternalDBf(rs.Error)
	}
	if rs.RowsAffected == 0 {
		return errores.NewNotFoundf(nil, errores.ErrRecordNotFaund)
	}
	return nil
}
func (r *collection) SaveTemplateDetails(tx *gorm.DB, id, username, tmpluuid, jsond string) error {
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
