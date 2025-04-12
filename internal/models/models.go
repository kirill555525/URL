package models

type Request struct {
	Url string `json:"url"`
}

type Response struct {
	ShortUrl string `json:"result"`
}
