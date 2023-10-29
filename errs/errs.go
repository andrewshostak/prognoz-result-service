package errs

type AliasNotFoundError struct {
	Message string
}

func (e AliasNotFoundError) Error() string {
	return e.Message
}

type MatchNotFoundError struct {
	Message string
}

func (e MatchNotFoundError) Error() string {
	return e.Message
}
