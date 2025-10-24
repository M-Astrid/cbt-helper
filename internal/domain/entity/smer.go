package entity

import (
	"time"
)

type SMEREntry struct {
	ID           string
	UserID       int64
	CreatedTime  time.Time
	UpdatedTime  time.Time
	Trigger      *Trigger
	Emotions     []*Emotion
	Thoughts     []*Thought
	Unstructured *Unstructured
}

func NewSMEREntry(userID int64) *SMEREntry {
	return &SMEREntry{
		ID:          "",
		UserID:      userID,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
}
