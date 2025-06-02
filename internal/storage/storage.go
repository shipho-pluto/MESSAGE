package storage

import "errors"

var (
	ErrMessageNotExist = errors.New("message does not exist")
	ErrNoMessagesFound = errors.New("no messages found")
	Banned             = errors.New("banned")
)
