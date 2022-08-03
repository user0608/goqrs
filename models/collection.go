package models

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID           string `chk:"uuid" gorm:"primaryKey" json:"id"`
	CollectionID string `chk:"uuid" json:"-"`
	Name         string `chk:"nonil max=80" json:"name"`
	Value        string `chk:"nonil max=480" json:"value"`
}
type Collection struct {
	ID              string         `chk:"uuid" gorm:"primaryKey" json:"id"`
	Name            string         `chk:"nonil" json:"name"`
	Description     string         `json:"description"`
	TimeOut         *time.Time     `json:"time_out"`
	NotBefore       *time.Time     `json:"not_before"`
	CreatedAt       time.Time      `json:"created_at"`
	NumTickets      int            `json:"num_tickets"`
	ImageTemplate   string         `json:"tamplate"`
	TemplateDetails string         `json:"template_details"`
	DocumentProcess string         `json:"document_process"`
	ProcessResult   string         `json:"process_result"`
	Tags            []Tag          `json:"tags"`
	AccountUsername string         `json:"-"`
	DeletedAt       gorm.DeletedAt `json:"-"`
}

type TemlateDetails struct {
	ItemWidth int `json:"item_width" form:"item_width"`
	QqSize    int `json:"qr_size" form:"qr_size"`
	QqXPos    int `json:"qr_x" form:"qr_x"`
	QqYPos    int `json:"qr_y" form:"qr_y"`
}
