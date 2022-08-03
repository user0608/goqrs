package models

type Ticket struct {
	ID           string `json:"id"`
	Reclaimed    string `json:"reclamed"`
	Annulled     string `json:"annulled"`
	CollectionID string `json:"-"`
}
