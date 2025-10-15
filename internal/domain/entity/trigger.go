package entity

type Trigger struct {
	Description string
}

func NewTrigger(description string) *Trigger {
	return &Trigger{
		Description: description,
	}
}
