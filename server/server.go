package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	"github.com/tosuke/hokuchi/flatcar"
	"github.com/tosuke/hokuchi/slogerr"
	"github.com/tosuke/hokuchi/storage"
)

type Server struct {
	Logger     *slog.Logger
	AssetsPath string
	Flatcar    *flatcar.Fetcher
	Storage    storage.Storage

	serv *http.Server
}

func (s *Server) HTTPHandler() http.Handler {
	logger := s.Logger
	if logger == nil {
		logger = slog.Default()
	}

	r := chi.NewRouter()
	r.Use(slogchi.New(logger.WithGroup("http")))
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Write([]byte("hello"))
	})
	r.HandleFunc("/boot_{arch}.efi", s.HandleBootbin)
	r.Get("/boot.ipxe", s.HandleBootstrapIPXE)
	r.Get("/ipxe", s.HandleIPXE)
	r.Get("/flatcar/{channel}/{arch}/{version}/kernel", s.HandleFlatcarKernel)
	r.Get("/flatcar/{channel}/{arch}/{version}/version", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		channel := chi.URLParam(r, "channel")
		arch := chi.URLParam(r, "arch")
		version := chi.URLParam(r, "version")

		key, err := s.Flatcar.ResolveKey(ctx, channel, arch, version)
		if err != nil {
			slog.ErrorContext(ctx, "resolve key", slogerr.Err(err))
			return
		}

		fmt.Fprintf(w, "%s", key.String())

	})

	return r
}
