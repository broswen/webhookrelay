package retry

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRetry(t *testing.T) {
	counter := 0
	r := NewRetry(time.Millisecond, 3, func() (any, error, bool) {
		counter += 1
		return nil, errors.New("test error"), true
	})
	_, err := r()
	assert.Equal(t, 4, counter)
	assert.EqualError(t, err, "test error")
}

func TestNewRetry_Succeeds(t *testing.T) {
	counter := 0
	r := NewRetry(time.Millisecond, 3, func() (any, error, bool) {
		counter += 1
		if counter > 1 {
			return nil, nil, false
		}
		return nil, errors.New("test error"), true
	})
	_, err := r()
	assert.Equal(t, 2, counter)
	assert.NoError(t, err)
}

func TestNewRetry_NotRetryable(t *testing.T) {
	counter := 0
	r := NewRetry(time.Millisecond, 3, func() (any, error, bool) {
		counter += 1
		return nil, errors.New("test error"), false
	})
	_, err := r()
	assert.Equal(t, 1, counter)
	assert.EqualError(t, err, "test error")
}
