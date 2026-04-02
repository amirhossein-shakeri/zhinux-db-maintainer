package shared

// TODO: Move to `../zhinux-platform/` as a shared standard?
type Pagination struct {
	Limit  int
	Offset int

	// todo: Modern? Which one is best practice + most use case coverage?
	Page     int
	PageSize int

	Total      int
	TotalPages int
}

type Result[T any] struct {
	Pagination Pagination
	Items      []T
	Total      int
}
