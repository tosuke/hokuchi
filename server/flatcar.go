package server

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tosuke/hokuchi/slogerr"
	"github.com/tosuke/hokuchi/storage"
)

func (s *Server) HandleFlatcarKernel(w http.ResponseWriter, r *http.Request) {
	channel := chi.URLParam(r, "channel")
	arch := chi.URLParam(r, "arch")
	version := chi.URLParam(r, "version")

	ctx := r.Context()
	w.Header().Set("Content-Type", "application/octet-stream")

	key, err := s.Flatcar.ResolveKey(ctx, channel, arch, version)
	if err != nil {
		slog.ErrorContext(ctx, "failed to resolve key", slogerr.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	size, reader, err := s.Storage.Get(ctx, key.String())
	if err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		io.Copy(w, reader)
		return
	} else if err != storage.ErrNotfound {
		slog.ErrorContext(ctx, "failed to get kernel from storage: %w", slogerr.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := s.Storage.Add(ctx, key.String())
	if err != nil {
		slog.ErrorContext(ctx, "failed to create tx for kernel: %w", slogerr.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var buf bytes.Buffer
	writer := io.MultiWriter(&buf, tx)
	if err := s.Flatcar.FetchKernel(ctx, writer, key); err != nil {
		slog.ErrorContext(ctx, "failed to fetch kernel", slogerr.Err(err))
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to commit kernel data: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, "kernel", time.Now(), bytes.NewReader(buf.Bytes()))
}
