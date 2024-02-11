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

type ErrInvalidRequest struct {
	Err error
}

func (e ErrInvalidRequest) Error() string {
	return fmt.Sprintf("invalid request: %s", e.Err.Error())
}

func (e ErrInvalidRequest) Unwrap() error {
	return e.Err
}
