package server

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/tosuke/hokuchi"
)

func (s *Server) HandleBootbin(w http.ResponseWriter, r *http.Request) {
	arch := hokuchi.NormalizeArch(chi.URLParam(r, "arch"))

	switch arch {
	case "arm64":
		http.ServeFile(w, r, filepath.Join(s.AssetsPath, "boot_arm64.efi"))
	case "amd64":
		http.ServeFile(w, r, filepath.Join(s.AssetsPath, "boot_amd64.efi"))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
