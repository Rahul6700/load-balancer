package models

type ServerInput struct {
	URL string `json:"url"`
	Port int  `json:"port"`
	Active int `json:"active"`
}
