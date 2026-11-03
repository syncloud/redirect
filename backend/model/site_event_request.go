package model

type SiteEventRequest struct {
	Event    string `json:"event"`
	Gclid    bool   `json:"gclid"`
	Language string `json:"language"`
}
