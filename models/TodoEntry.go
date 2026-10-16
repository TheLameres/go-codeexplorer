package models

import "time"

type TodoEntry struct {
	Id          int64
	CreatedAt   time.Time
	Description string
	Done        bool
}
