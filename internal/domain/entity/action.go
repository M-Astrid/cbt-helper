package entity

type Action struct {
	Description string
}

func NewAction(description string) *Action {
	return &Action{
		Description: description,
	}
}
