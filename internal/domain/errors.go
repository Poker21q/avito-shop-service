package domain

import (
	"errors"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInternalServerError = errors.New("internal server error")
	ErrNotFound            = errors.New("not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
)
