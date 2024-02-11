package repository

import "fmt"

type ErrUnknown struct {
	Err error
}

func (e ErrUnknown) Error() string {
	return e.Err.Error()
}

func (e ErrUnknown) Unwrap() error {
	return e.Err
}

type ErrWebhookNotFound struct {
	id string
}

func (e ErrWebhookNotFound) Error() string {
	return fmt.Sprintf("webhook not found: %s", e.id)
}

func (e ErrWebhookNotFound) Unwrap() error {
	return nil
}

type ErrInvalidData struct {
	Message string
}

func (e ErrInvalidData) Error() string {
	return e.Message
}
