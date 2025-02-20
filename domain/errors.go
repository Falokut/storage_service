package domain

import (
	"errors"
)

var (
	ErrFileNotFound = errors.New("file not found")
)

const (
	ErrCodeFileNotFound        = 600
	ErrCodeFileTooBig          = 601
	ErrCodeFileHasZeroSize     = 602
	ErrCodeUnsupportedFileType = 603
	ErrCodeInvalidRange        = 604
)

type InvalidArgumentError struct {
	ErrCode int
	Reason  string
}

func NewInvalidArgumentError(reason string, errCode int) InvalidArgumentError {
	return InvalidArgumentError{Reason: reason, ErrCode: errCode}
}

func (e InvalidArgumentError) Error() string {
	return e.Reason
}
