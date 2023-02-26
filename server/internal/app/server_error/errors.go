package server_error

import "errors"

var (
	ErrPreviewNotFound = errors.New("some preview not found")
	ErrEmptyUrlSl      = errors.New("UrlSL is empty slice")
	ErrUnknown         = errors.New("unknown error on server")
)
