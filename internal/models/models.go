package models

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortURL string `json:"result"`
}

type RequestBatchUrl struct {
	CorrelationId string `json:"correlation_id"`
	OriginalUrl   string `json:"original_url"`
}

type ResponseBatchUrl struct {
	CorrelationId string `json:"correlation_id"`
	ShortUrl      string `json:"short_url"`
}
