package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
)

type Server struct {
	Logger *slog.Logger
    AssetsPath string
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
		w.Write([]byte("hello"))
	})
    r.HandleFunc("/boot_{arch}.efi", s.HandleBootbin)
    r.Get("/boot.ipxe", s.HandleBootstrapIPXE)

	return r
}

func (s *Server) Start(ctx context.Context, addr string) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	hs := &http.Server{
		Addr:    addr,
		Handler: s.HTTPHandler(),
	}
	go func() {
		if err := hs.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			cancel(fmt.Errorf("failed to listen HTTP server: %w", err))
		}
	}()

	<-ctx.Done()
	if err := context.Cause(ctx); err != nil && err != context.Canceled {
		return err
	}

	// graceful shutdown
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	return hs.Shutdown(ctx)
}

/*
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				stack_trace := fmt.Sprintf("%v\n%s", rvr, buf)
				slog.ErrorContext(r.Context(), "panic", slog.String("stack_trace", stack_trace))

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})

}
*/
