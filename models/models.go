package models

type ServerStruct struct {
	URL string `json:"url"`
	//Port int  `json:"port"`
	Active int `json:"active"`
	Index int //needed for heap
}
