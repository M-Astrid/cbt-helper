package domainError

type ValidationError struct {
	Msg string
}

func (ve ValidationError) Error() string {
	return ve.Msg
}
