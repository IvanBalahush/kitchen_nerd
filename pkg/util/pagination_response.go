package util

type PaginationResponse struct {
	Size  uint64 `json:"size"`
	Page  uint64 `json:"page"`
	Total uint64 `json:"total"`
}

func NewPaginationResponse(size, page, total uint64) *PaginationResponse {
	return &PaginationResponse{Size: size, Page: page, Total: total}
}
