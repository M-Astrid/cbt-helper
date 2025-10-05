package entity

import "time"

type SMEREntry struct {
	ID          int64
	CreatedTime time.Time
}

func NewSMEREntry() *SMEREntry {
	return &SMEREntry{
		ID:          0,
		CreatedTime: time.Now(),
	}
}
