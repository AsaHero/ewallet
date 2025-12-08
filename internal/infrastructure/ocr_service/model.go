package ocr_service

type ErrorResponse struct {
	Details string `json:"details"`
}

type ImageToTextResponse struct {
	Lines []struct {
		Text       string  `json:"text"`
		Confidence float64 `json:"confidence"`
		Box        [][]int `json:"box"`
	} `json:"lines"`
	FullText string `json:"full_text"`
}
