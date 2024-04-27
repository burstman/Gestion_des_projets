package data

import "errors"

var (
	ErrNoRecord        = errors.New("data: no matching data found")
	ErrDuplicateRecord = errors.New("data: duplicate data found")
	ErrDuplicateEmail  = errors.New("data: duplicate email found")
	ErrDuplicateName   = errors.New("data: duplicate name found")
	ErrInvalidCredentials = errors.New("data: invalid credentials")
)
