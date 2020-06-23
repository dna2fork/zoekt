package web

import (
	"net/http"
	"bytes"

	"github.com/google/zoekt/contrib/keyval"
)

func (s *Server) serveKeyval(w http.ResponseWriter, r *http.Request) {
	if !s.BasicAuth.checkAuth(r) {
		w.WriteHeader(401)
		w.Write(bytes.NewBufferString("Not authenticated.").Bytes())
		return
	}
	keyval.StorageHandler(w, r)
}
