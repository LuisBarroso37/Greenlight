package data

import (
	"math"
	"strings"

	"github.com/LuisBarroso37/Greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string // Holds the supported sort values
}

// Validate filters received as query parameters
func ValidateFilters(v *validator.Validator, filters Filters) {
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filters.PageSize <= 100, "page_size", " must be a maximum of 100")
	v.Check(validator.In(filters.Sort, filters.SortSafelist...), "sort", "invalid sort value")
}

// Check that the client-provided `Sort` field matches one of the entries in our safelist
// and if it does, extract the column name from the Sort field by stripping the leading
// hyphen character (if one exists).
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

// Return the sort direction ("ASC" or "DESC") depending on the prefix character of the `Sort` field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

// Return the number of records to be returned in the query
func (f Filters) limit() int {
	return f.PageSize
}

// Return the number of rows to skip before starting to return records from the query
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// Struct used for holding the pagination metadata
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// Calculates the appropriate pagination metadata values
// given the total number of records, current page, and page size values
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
