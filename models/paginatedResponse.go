package models

type PaginatedResponse struct {
	Items       []Item `json:"items"`
	TotalItems  int    `json:"totalItems"`
	TotalPages  int    `json:"totalPages"`
	CurrentPage int    `json:"currentPage"`
}
