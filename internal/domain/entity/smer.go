package entity

import (
	"time"

	"github.com/google/uuid"
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
		ID:          uuid.New().String(),
		UserID:      userID,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
}
