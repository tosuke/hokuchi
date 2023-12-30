package server

import (
	"log/slog"
	"net/http"
	"time"

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
