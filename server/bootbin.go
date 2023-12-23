package server

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func (s *Server) HandleBootbin(w http.ResponseWriter, r *http.Request) {
	arch := chi.URLParam(r, "arch")

	switch arch {
	case "aarch64", "arm64":
		http.ServeFile(w, r, filepath.Join(s.AssetsPath, "boot_aarch64.efi"))
	case "x86_64", "amd64", "intel64":
		http.ServeFile(w, r, filepath.Join(s.AssetsPath, "boot_x86_64.efi"))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
