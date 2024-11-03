package repository

import "errors"

var (
	ErrInvalidName   = errors.New("invalid name")
	ErrNotFound      = errors.New("not found")
	ErrStillInUse    = errors.New("still used")
	ErrSQL           = errors.New("unhandled sql error")
	ErrDuplicateName = errors.New("duplicate name")
	ErrValidation    = errors.New("validation error")
	ErrTooSmall      = errors.New("number too small")
	ErrTooBig        = errors.New("number too big")
	ErrParseInt      = errors.New("string not containing a number")
)
