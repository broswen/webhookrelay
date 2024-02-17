package retry

import (
	"math"
	"time"
)

// NewRetry returns a higher order function that retries f a maximum of maxRetries with exponential backoff starting with
// delay time. It will only retry if f returns b == true, meaning the error is retryable
func NewRetry[T any](delay time.Duration, maxRetries int, f func() (T, error, bool)) func() (T, error) {
	return func() (T, error) {
		attempts := 0
		var t T
		var err error
		var b bool
		t, err, b = f()
		for err != nil && attempts < maxRetries && b {
			time.Sleep(delay * time.Duration(math.Pow(2.0, float64(attempts))))
			attempts += 1
			t, err, b = f()
		}
		return t, err
	}
}
