package service

import "fmt"

type ErrTokenInProgress struct {
	token string
}

func (e ErrTokenInProgress) Error() string {
	return fmt.Sprintf("request in progress: %s", e.token)
}

func (e ErrTokenInProgress) Unwrap() error {
	return nil
}
