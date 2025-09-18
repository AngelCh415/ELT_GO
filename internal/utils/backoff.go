package utils

import (
	"time"
)

type Backoff struct {
	base       time.Duration
	maxRetries int
}

func NewBackoff(base time.Duration, maxRetries int) Backoff {
	return Backoff{base: base, maxRetries: maxRetries}
}

func (b Backoff) Do(fn func(i int) error) error {
	var err error
	for i := 0; i <= b.maxRetries; i++ {
		err = fn(i)
		if err == nil {
			return nil
		}
		t := time.Duration(1<<i) * b.base
		time.Sleep(t)
	}
	return err
}
