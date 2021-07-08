// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"path/filepath"

	"golang.org/x/net/context"

	"github.com/google/zoekt/keyval"
)

const defaultNumResults = 50

type Server struct {
	// Searcher zoekt.Searcher

	// Serve HTML interface
	HTML bool

	// If set, show files from the index.
	Print bool

	// Version string for this server.
	Version string

	// Depending on the Host header, add a query to the entry
	// page. For example, when serving on "search.myproject.org"
	// we could add "r:myproject" automatically.  This allows a
	// single instance to serve as search engine for multiple
	// domains.
	HostCustomQueries map[string]string

	startTime time.Time

	lastStatsMu sync.Mutex
	lastStatsTS time.Time

	SourceBaseDir string
	IndexDir string
	BasicAuth ServerAuthBasic
}

func NewMux(s *Server) (*http.ServeMux, error) {
	s.startTime = time.Now()

	mux := http.NewServeMux()

	if s.HTML {
		mux.HandleFunc("/search", s.serveSearch)
		mux.HandleFunc("/", s.serveHome)
		mux.HandleFunc("/about", s.serveAbout)
	}
	if s.HTML && s.SourceBaseDir != "" {
		var err error
		s.SourceBaseDir, err = filepath.Abs(s.SourceBaseDir)
		if err != nil {
			log.Fatal(fmt.Sprintf("set invalid source code base directory of \"%s\": %v", s.SourceBaseDir, err))
		}
		if isDirectory(s.SourceBaseDir) != 1 {
			log.Fatal(fmt.Sprintf("source code base directory of \"%s\" is not a directory or inaccessible", s.SourceBaseDir))
		}
		mux.HandleFunc("/fsprint", s.serveFSPrint)
		mux.HandleFunc("/scmprint", s.serveScmPrint)
	}
	if keyval.IsKeyvalFSEnabled() {
		log.Printf("[kv] key-value service is running ...,")
		mux.HandleFunc("/keyval", s.serveKeyval)
	}

	return mux, nil
}

func (s *Server) serveSearch(w http.ResponseWriter, r *http.Request) {
	if !s.BasicAuth.checkAuth(r) {
		w.WriteHeader(401)
		w.Write(bytes.NewBufferString("Not authenticated.").Bytes())
		return
	}
	err := s.serveSearchErr(w, r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveSearchErr(w http.ResponseWriter, r *http.Request) error {
	qvals := r.URL.Query()
	queryStr := qvals.Get("q")
	if queryStr == "" {
		return fmt.Errorf("no query found")
	}
	numStr := qvals.Get("num")
	num, err := strconv.Atoi(numStr)

	if err != nil || num <= 0 {
		num = defaultNumResults
	}

	w.Write([]byte(`{"error":"not implemented yet"}`))
	return nil
}

const statsStaleNess = 30 * time.Second

func (s *Server) fetchStats(ctx context.Context) (int, error) {
	s.lastStatsMu.Lock()
	// stats := s.lastStats
	if time.Since(s.lastStatsTS) > statsStaleNess {
		// stats = nil
	}
	s.lastStatsMu.Unlock()

	// TDO: check status

	s.lastStatsMu.Lock()
	s.lastStatsTS = time.Now()
	// s.lastStats = stats
	s.lastStatsMu.Unlock()

	return 0, nil
}

func (s *Server) serveHomeErr(w http.ResponseWriter, r *http.Request) error {
	_, err := s.fetchStats(r.Context())
	if err != nil {
		return err
	}
	w.Write([]byte(`{"error":"not implemented yet"}`))
	return nil
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	if err := s.serveHomeErr(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveAboutErr(w http.ResponseWriter, r *http.Request) error {
	_, err := s.fetchStats(r.Context())
	if err != nil {
		return err
	}

	w.Write([]byte(`{"error":"not implemented yet"}`))
	return nil
}

func (s *Server) serveAbout(w http.ResponseWriter, r *http.Request) {
	if err := s.serveAboutErr(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}
