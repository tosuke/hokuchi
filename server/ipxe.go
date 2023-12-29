package server

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/tosuke/hokuchi"
	"github.com/tosuke/hokuchi/slogerr"
)

const bootstrapIpxe = `#!ipxe
chain ipxe?uuid=${uuid}&mac=${mac:hexhyp}&domain=${domain}&hostname=${hostname}&serial=${serial}&arch=${buildarch:uristring}
`

var ipxeTemplate = template.Must(template.New("ipxe").Parse(`#!ipxe
kernel {{.Kernel.URI}}{{range $arg := .Kernel.Args}} {{$arg}}{{end}}
{{- range $image := .Images }}
initrd {{if ne $image.Name ""}} --name {{$image.Name}}{{end}} {{$image.URI}}
{{- end}}
boot`))

func (s *Server) HandleBootstrapIPXE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(bootstrapIpxe))
}

func (s *Server) HandleIPXE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dummyProfile := &hokuchi.Profile{
		ID:     "dummy",
		Arch:   "arm64",
		Labels: nil,
		BootConfig: hokuchi.BootConfig{
			Kernel: hokuchi.Kernel{
				URI: "http://beta.release.flatcar-linux.net/arm64-usr/3760.1.1/flatcar_production_pxe.vmlinuz",
				Args: []string{
					"initrd=main",
					"flatcar.first_boot=1",
					"console=tty0",
					"flatcar.autologin=tty0",
				},
			},
			Images: []hokuchi.Image{
				{
					Name: "main",
					URI:  "http://beta.release.flatcar-linux.net/arm64-usr/3760.1.1/flatcar_production_pxe_image.cpio.gz",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "text/plain")
	var buf bytes.Buffer
	if err := ipxeTemplate.Execute(&buf, dummyProfile.BootConfig); err != nil {
		slog.ErrorContext(ctx, "Error rendering template", slogerr.Err(err))
		renderIPXEError(w, http.StatusInternalServerError, "")
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		slog.ErrorContext(ctx, "Error writing response", slogerr.Err(err))
		renderIPXEError(w, http.StatusInternalServerError, "")
		return
	}
}

func renderIPXEError(w http.ResponseWriter, code int, message string) {
	if message == "" {
		message = http.StatusText(code)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	fmt.Fprintf(w, "#!ipxe\necho %d - %s\nshell\n", code, message)
}
