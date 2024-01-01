package server

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tosuke/hokuchi/slogerr"
	"github.com/tosuke/hokuchi/storage"
)

func (s *Server) HandleFlatcarKernel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	profile := defaultProfile

	fc := profile.Boot.Flatcar
	if fc == nil {
		http.Error(w, "profile does not use flatcar", http.StatusBadRequest)
		return
	}

	key, err := s.Flatcar.ResolveKey(ctx, fc.Channel, profile.Arch, fc.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Error resolving flatcar version", slogerr.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	size, reader, err := s.Storage.Get(ctx, key.KernelKey())
	if err != nil {
		if errors.Is(err, storage.ErrNotfound) {
			http.Error(w, fmt.Sprintf("flatcar kernel %s not found", key.String()), http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "Error getting kernel from storage", slogerr.Err(err))
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Content-Disposition", "kernel")
	if _, err := io.Copy(w, reader); err != nil {
		slog.ErrorContext(ctx, "Error writing kernel response", slogerr.Err(err))
		return
	}
	if err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		io.Copy(w, reader)
		return
	} else if !errors.Is(err, storage.ErrNotfound) {
		slog.ErrorContext(ctx, "Error getting kernel from storage", slogerr.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleFlatcarInitrd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profile := defaultProfile
	fc := profile.Boot.Flatcar
	if fc == nil {
		http.Error(w, "profile does not use flatcar", http.StatusBadRequest)
		return
	}

	key, err := s.Flatcar.ResolveKey(ctx, fc.Channel, profile.Arch, fc.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Error resolving flatcar version", slogerr.Err(err))
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	size, reader, err := s.Storage.Get(ctx, key.InitrdKey())
	if err != nil {
		if errors.Is(err, storage.ErrNotfound) {
			http.Error(w, fmt.Sprintf("flatcar initrd %s not found", key.String()), http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "Error getting initrd from storage", slogerr.Err(err))
		http.Error(w, "Internal Error from storage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Content-Disposition", "initrd")
	if _, err := io.Copy(w, reader); err != nil {
		slog.ErrorContext(ctx, "Error writing initrd response", slogerr.Err(err))
		return
	}
}
