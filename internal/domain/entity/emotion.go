package entity

import (
	"fmt"
	"strconv"

	domainError "github.com/M-Astrid/cbt-helper/internal/domain/error"
)

type Emotion struct {
	Name  string
	Scale int
}

func NewEmotion(name string, scale string) (*Emotion, error) {
	intScale, err := strconv.Atoi(scale)
	if err != nil {
		err = domainError.ValidationError{Msg: "Emotion scale is not a digit"}
	}
	if intScale < 0 || intScale > 100 {
		err = domainError.ValidationError{Msg: fmt.Sprintf("%s scale is not in range [0, 10]", name)}
	}
	return &Emotion{
		Name:  name,
		Scale: intScale,
	}, err
}
