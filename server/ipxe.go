package server

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"text/template"

	"braces.dev/errtrace"
	"github.com/tosuke/hokuchi/profile"
	"github.com/tosuke/hokuchi/slogerr"
)

const bootstrapIpxe = `#!ipxe
chain --replace ipxe?uuid=${uuid}&mac=${mac:hexhyp}&domain=${domain}&hostname=${hostname}&serial=${serial}&arch=${buildarch:uristring}
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

var defaultProfile = profile.Profile{
	ID: "default",
	Arch: "arm64",
	Boot: profile.Boot{
		Flatcar: &profile.Flatcar{
			Channel: "beta",
			Version: "current",
			Args: []string{
				"flatcar.firstboot=1",
				"sshkey=\"ecdsa-sha2-nistp384 AAAAE2VjZHNhLXNoYTItbmlzdHAzODQAAAAIbmlzdHAzODQAAABhBBW8gye009QtUJvF6by0Cmx91y0jRA5h0FGzurmazR6w9MSW702YFM/cXgsf3Au4y6kzRmUSpUDdm7QhRYM4yDsljmoyLZ2qz807sCakOaFhXmYrFeqbgQy2RveSeVT94A== tosuke\"",
			},
		},
	},
}

const backoffBase = 10_000
const backoffCap = 600_000

func (s *Server) HandleIPXE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	var attempt int
	if q := query.Get("attempt"); q != "" {
		if v, err := strconv.ParseInt(q, 10, 0); err == nil {
			attempt = int(v)
		}
	}

	// retry
	// Exponential Backoff And Equal Jitter
	temp := min(backoffCap, backoffBase*int(math.Pow(2, float64(attempt))))
	sleepMs := temp/2 + rand.Intn(temp/2)
	slog.Debug("retry", slog.Int("sleep", sleepMs))
	sleep := sleepMs / 1000
	if err := renderRetryIPXE(w, r, sleep, attempt+1); err != nil {
		slog.ErrorContext(ctx, "Error writing retry ipxe response", slogerr.Err(err))
	}
	return
}

func renderRetryIPXE(w http.ResponseWriter, r *http.Request, sleep int, nextAttempt int) error {
	w.Header().Set("Content-Type", "text/plain")

	retryURL := *r.URL
	query := retryURL.Query()
	query.Set("attempt", strconv.FormatInt(int64(nextAttempt), 10))
	retryURL.RawQuery = query.Encode()

	var b bytes.Buffer
	fmt.Fprintln(&b, "#!ipxe")
	fmt.Fprintf(&b, "echo Retry after %ds...\n", sleep)
	fmt.Fprintf(&b, "sleep %d\n", sleep)
	fmt.Fprintf(&b, "chain --replace %s\n", retryURL.String())

	if _, err := b.WriteTo(w); err != nil {
		return errtrace.Wrap(err)
	}

	return nil
}

func renderIPXEError(w http.ResponseWriter, code int, message string) error {
	if message == "" {
		message = http.StatusText(code)
	}
	var b bytes.Buffer
	fmt.Fprintln(&b, "#!ipxe")
	fmt.Fprintf(&b, "echo %d - %s\n", code, message)
	fmt.Fprintln(&b, "shell")

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)

	if _, err := b.WriteTo(w); err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}
