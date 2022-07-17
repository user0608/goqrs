package models

type Tiket struct {
	ID           string `json:"id"`
	Reclaimed    string `json:"reclamed"`
	Annulled     string `json:"annulled"`
	CollectionID string `json:"-"`
}
