package service

import "errors"

var (
	errEmptyURL      = errors.New("Empty URL")
	errUnparsableURL = errors.New("Both the host and path fields are empty")
)
