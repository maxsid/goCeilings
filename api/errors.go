package api

import "errors"

var (
	ErrCouldNotExistURLVariable   = errors.New("could not exist URL variable")
	ErrCouldNotReadUserIDFromCtx  = errors.New("could not read user.ID from context")
	ErrCouldNotReadDrawingFromCtx = errors.New("could not read drawing from context")
	ErrUserNotFound               = errors.New("user not found")
	ErrUserAlreadyExist           = errors.New("user with this login already exist")
	ErrDrawingNotFound            = errors.New("drawing not found")
)
