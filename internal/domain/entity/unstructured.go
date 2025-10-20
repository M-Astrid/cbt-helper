package entity

type Unstructured struct {
	Text string
}

func NewUnstructured(description string) *Unstructured {
	return &Unstructured{
		Text: description,
	}
}
