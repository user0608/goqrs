package models

import "time"

type Tag struct {
	ID           string `json:"id"`
	CollectionID string `json:"-"`
	Name         string `json:"name"`
	Value        string `json:"value"`
}
type Collection struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	TimeOut         *time.Time `json:"time_out"`
	NotBefore       time.Time  `json:"not_before"`
	NumTickets      int        `json:"num_tickets"`
	AccountUsername string     `json:"-"`
	Tags            []Tag      `json:"tags"`
}
