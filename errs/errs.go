package errs

import "errors"

var (
	ErrIncorrectFixtureStatus          = errors.New("incorrect fixture status")
	ErrUnexpectedAPIFootballStatusCode = errors.New("unexpected status code received from api-football")
)

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

type UnexpectedNumberOfItemsError struct {
	Message string
}

func (e UnexpectedNumberOfItemsError) Error() string {
	return e.Message
}
