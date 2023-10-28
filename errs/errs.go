package errs

type AliasNotFoundError struct {
	Message string
}

func (e AliasNotFoundError) Error() string {
	return e.Message
}
