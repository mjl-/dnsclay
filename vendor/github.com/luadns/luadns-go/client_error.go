package luadns

import (
	"strconv"
	"strings"
)

// ErrBadStatusCode represents an error for unexpected status code returned by the API server.
type ErrBadStatusCode struct {
	StatusCode int
}

func (e *ErrBadStatusCode) Error() string {
	return "Server returned bad status code (" + strconv.Itoa(e.StatusCode) + ")"
}

// ErrBadContentType represents an error for unexpected content type returned by the API server.
type ErrBadContentType struct {
	ContentType string
}

func (e *ErrBadContentType) Error() string {
	return "Server returned bad content type (" + e.ContentType + ")"
}

// ErrTooManyRequests represents an error returned when users exeeds requests quota (status code 429).
type ErrTooManyRequests struct {
	Limit int64
	Reset int64
}

func (e *ErrTooManyRequests) Error() string {
	return "Too many requests, retry after " + strconv.FormatInt(e.Reset, 10) + " unix time"
}

// InputError represents a validation error returned by the API server
// when input data is invalid (status code 400).
type InputError struct {
	Classification string   `json:"classification"` // DeserializationError, RequiredError, ValidationError
	FieldNames     []string `json:"fieldNames"`     // Example: name
	Message        string   `json:"message"`        // Example: invalid name
}

func (e *InputError) Error() string {
	if len(e.FieldNames) == 0 {
		return e.Message
	}
	return "Invalid data for " + strings.Join(e.FieldNames, ", ") + ": " + e.Message
}

// BadRequestError represents a list of validation errors returned by the API server.
type BadRequestError []InputError

func (e BadRequestError) Error() string {
	errs := []string{}
	for _, err := range e {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, "; ")
}

// ForbiddenRequestError represents an error returned by the server when input
// data is valid but the operation is not allowed (status code 403).
type ForbiddenRequestError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e *ForbiddenRequestError) Error() string {
	return e.Status + ": " + e.Message
}
