package models

import "time"

type Ticket struct {
	ID           string      `json:"id"`
	Reclaimed    *time.Time  `json:"reclamed"`
	Annulled     *time.Time  `json:"annulled"`
	CollectionID string      `json:"-"`
	Collection   *Collection `json:"collection" gorm:"-"`
}
