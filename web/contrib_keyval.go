package web

import (
	"net/http"
	"net/url"
	"bytes"
	"path/filepath"
	"strconv"

	"github.com/google/zoekt/contrib"
	"github.com/google/zoekt/keyval"
)

func (s *Server) serveKeyval(w http.ResponseWriter, r *http.Request) {
	if !s.BasicAuth.checkAuth(r) {
		w.WriteHeader(401)
		w.Write(bytes.NewBufferString("Not authenticated.").Bytes())
		return
	}

	qv := r.URL.Query()
	cmd := qv.Get("a")
	switch cmd {
	case "search":
		// put keyval index into /storage/.zoekt/index
		// XXX: pretent as a p4 repo? easy to ignore .p4 folder
		indexPath := filepath.Join(keyval.GetBaseDir(), ".p4", ".zoekt", "index")
		s.contribSearchKeyval(indexPath, qv, w, r)
	default:
		keyval.ServeBasic(w, r)
	}
}

func (s *Server) contribSearchKeyval(indexPath string, keyval url.Values, w http.ResponseWriter, r *http.Request) {
	q := keyval.Get("q")
	n := keyval.Get("n")
	if q == "" {
		utilErrorStr(w, "empty query", 400)
		return
	}
	num, err := strconv.Atoi(n)
	if err != nil || num <= 0 {
		num = defaultNumResults
	}

	result, err := contrib.Search(indexPath, r.Context(), q, num)
	if err != nil {
		utilError(w, err, 500)
		return
	}

	s.contribRenderSearchResult(result, q, num, w, r)
}
