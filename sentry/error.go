package sentry

import (
	"errors"
)

// SentryErrors defines functions that a SentryError should provide
type SentryErrors interface {
	error
	AddExtra(key string, value interface{})
	AddTag(key, value string)
	GetTags() Tags
	GetExtras() Extras
	Unwrap() error
}

// Error wraps the stdlib error to make it possible to check if an error should be sent to Sentry
type Error struct {
	error
	tags   Tags
	extras Extras
}

// SentryError creates a new Error from an original error
func SentryError(err error) SentryErrors {
	if err == nil {
		return nil
	}

	return &Error{
		error:  err,
		tags:   Tags{},
		extras: Extras{},
	}
}

func (e Error) GetReason() error {
	return e.error
}

// AddTag adds a tag to be sent to Sentry
func (e *Error) AddTag(key, value string) {
	e.tags.Add(key, value)
}

func (e Error) GetTags() Tags {
	return e.tags
}

// AddExtra adds extra data to be sent to Sentry
func (e *Error) AddExtra(key string, value interface{}) {
	e.extras.Add(key, value)
}

func (e Error) GetExtras() Extras {
	return e.extras
}

func (e *Error) Unwrap() error {
	return errors.Unwrap(e.error)
}

// IsSentryError checks if a given error is of Error type and therefor should be sent to Sentry
func IsSentryError(err error) bool {
	_, ok := AsSentryError(err)

	return ok
}

// AsSentryError checks if a given error is of Error type or contains a wrapped error and returns it
func AsSentryError(err error) (*Error, bool) {
	var sentryError *Error
	ok := errors.As(err, &sentryError)

	return sentryError, ok
}
