package models

type PaginationRequest struct {
	Limit  uint64 `form:"limit" json:"limit"`
	Offset uint64 `form:"offset" json:"offset"`
}

type PaginationResponse struct {
	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
	Total  int64  `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
