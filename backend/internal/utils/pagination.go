package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// PaginationParams holds the structured output of parsing pagination.
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// GetPagination parses page and limit from the query string, enforcing reasonable bounds.
func GetPagination(c *gin.Context) PaginationParams {
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	if err != nil || page < 1 {
		page = DefaultPage
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))
	if err != nil || (limit < 1 && limit != -1) {
		limit = DefaultLimit
	}
	if limit > MaxLimit && limit != -1 {
		limit = MaxLimit
	}

	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
