package storage

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrDoesNotExist  = errors.New("does not exist")
	ErrAbort         = errors.New("abort")
)

func handleStorageError(err error) error {
	if err == gorm.ErrRecordNotFound {
		return ErrDoesNotExist
	}
	return err
}
