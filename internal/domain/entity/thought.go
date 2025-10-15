package entity

import "time"

type Thought struct {
	Description string
	CreatedTime time.Time
	UpdatedTime time.Time
}

func NewThought(d string) (*Thought, error) {
	return &Thought{
		Description: d,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}, nil
}
