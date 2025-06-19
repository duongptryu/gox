package pagination

// Paging represents pagination information
type Paging struct {
	Page       int   `json:"page" form:"page"`
	Limit      int   `json:"limit" form:"limit"`
	Total      int64 `json:"total" form:"total"`
	NextCursor int   `json:"next_cursor"`
}

// Fulfill applies default values to pagination parameters
func (p *Paging) Fulfill() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10
	}
}

// GetOffset calculates the database offset
func (p *Paging) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// HasNext checks if there is a next page
func (p *Paging) HasNext() bool {
	totalPages := (p.Total + int64(p.Limit) - 1) / int64(p.Limit)
	return int64(p.Page) < totalPages
}

// HasPrev checks if there is a previous page
func (p *Paging) HasPrev() bool {
	return p.Page > 1
}

// GetTotalPages calculates total pages
func (p *Paging) GetTotalPages() int64 {
	if p.Total == 0 {
		return 1
	}
	return (p.Total + int64(p.Limit) - 1) / int64(p.Limit)
}
