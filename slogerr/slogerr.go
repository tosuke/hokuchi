package slogerr

import "log/slog"

func Err(err error) slog.Attr {
	return slog.String("err", err.Error())
}
