package server

import "net/http"

const bootstrapIpxe = `#!ipxe
chain ipxe?uuid=${uuid}&mac=${mac:hexhyp}&domain=${domain}&hostname=${hostname}&serial=${serial}&arch=${buildarch:uristring}
`

func (s *Server) HandleBootstrapIPXE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(bootstrapIpxe))
}
