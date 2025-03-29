package schemas

type PageResponse[T any] struct {
	Total   uint `json:"total"`
	Page    uint `json:"page"`
	Limit   uint `json:"limit"`
	Entries []T  `json:"entries"`
}
