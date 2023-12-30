package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"braces.dev/errtrace"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	"github.com/tosuke/hokuchi/flatcar"
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
	// r.Get("/profile/{pid}/flatcar/kernel")
	// r.Get("/profile/{pid}/flatcar/initrd")
	// r.Get("/profile/{pid}/ignition")
	// TODO やめる
	r.Get("/flatcar/{channel}/{arch}/{version}/kernel", s.HandleFlatcarKernel)

	return r
}

func (s *Server) Start(addr string) error {
	if s.serv != nil {
		return errtrace.New("server already started")
	}
	s.serv = &http.Server{
		Addr:    addr,
		Handler: s.HTTPHandler(),
	}
	defer func() { s.serv = nil }()
	if err := s.serv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return errtrace.Wrap(err)
	}
	return nil
}

func (s *Server) Close() error {
	if s.serv == nil {
		return errtrace.New("server not started")
	}
	return errtrace.Wrap(s.serv.Close())
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.serv == nil {
		return errtrace.New("server not started")
	}
	return errtrace.Wrap(s.serv.Shutdown(ctx))
}
