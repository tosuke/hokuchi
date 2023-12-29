package slogerr

import (
	"log/slog"
	"strings"

	"braces.dev/errtrace"
)

func Err(err error) slog.Attr {
	msg := err.Error()
	trace := errtrace.FormatString(err)
	trace = strings.TrimPrefix(trace, msg)
	trace = strings.TrimSpace(trace)
	return slog.Group("err", slog.String("msg", msg), slog.String("trace", trace))
}
