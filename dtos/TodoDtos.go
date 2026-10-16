package dtos

import (
	"strings"
	"time"
)

type CreateNewTodoEntryRequest struct {
	Description string `json:"description"`
}

func (r CreateNewTodoEntryRequest) Validate() string {
	d := strings.TrimSpace(r.Description)
	switch {
	case d == "":
		return "description is required"
	case len(d) > 500:
		return "description is too long (max 500)"
	}
	return ""
}

type CreateNewTodoEntryResponse struct {
	Id          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
}

type TodoEntryResponse struct {
	Id          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
