package entity

type Trigger struct {
	ID          int64
	Description string
	SMERID      int64
}

func NewTrigger(description string, smerID int64) *Trigger {
	return &Trigger{
		Description: description,
		SMERID:      smerID,
	}
}
