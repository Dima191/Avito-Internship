package repository

import "errors"

var (
	ErrOpenConn = errors.New("failed to open connection to db")
	ErrPingDB   = errors.New("failed to ping db")
)
