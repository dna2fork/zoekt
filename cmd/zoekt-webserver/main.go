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

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/zoekt/web"
	"github.com/google/zoekt/contrib"
	"github.com/google/zoekt/analysis"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/net/trace"
)

const logFormat = "2006-01-02T15-04-05.999999999Z07"

func divertLogs(dir string, interval time.Duration) {
	t := time.NewTicker(interval)
	var last *os.File
	for {
		nm := filepath.Join(dir, fmt.Sprintf("zoekt-webserver.%s.%d.log", time.Now().Format(logFormat), os.Getpid()))
		fmt.Fprintf(os.Stderr, "writing logs to %s\n", nm)

		f, err := os.Create(nm)
		if err != nil {
			// There is not much we can do now.
			fmt.Fprintf(os.Stderr, "can't create output file %s: %v\n", nm, err)
			os.Exit(2)
		}

		log.SetOutput(f)
		last.Close()

		last = f

		<-t.C
	}
}

const (
	Version = "pre-kill-zoekt"
)

func main() {
	logDir := flag.String("log_dir", "", "log to this directory rather than stderr.")
	logRefresh := flag.Duration("log_refresh", 24*time.Hour, "if using --log_dir, start writing a new file this often.")

	listen := flag.String("listen", ":20220", "listen on this address.")
	html := flag.Bool("html", true, "enable HTML interface")
	fsbase := flag.String("fs_base_dir", "", "enable api to fetch file/directory contents (filepath)")
	basicauth := flag.String("basic_auth", "", "enable basic auth in api invocation (filepath)")
	enablePprof := flag.Bool("pprof", false, "set to enable remote profiling.")
	sslCert := flag.String("ssl_cert", "", "set path to SSL .pem holding certificate.")
	sslKey := flag.String("ssl_key", "", "set path to SSL .pem holding key.")
	hostCustomization := flag.String(
		"host_customization", "",
		"specify host customization, as HOST1=QUERY,HOST2=QUERY")
	version := flag.Bool("version", false, "Print version number")
	flag.Parse()

	if contrib.DEBUG_ON {
		fmt.Println("[!] Debug Mode: ON")
	}
	if *fsbase == "" {
		fmt.Println("[!] No source root (--fs_base_dir)")
	}
	if analysis.CYGWIN_ON {
		fmt.Println("[!] use git/p4 in CygWin, CygWin base directory:", analysis.CYGWIN_BASE_DIR)
	}
	if analysis.P4_BIN == "" {
		fmt.Println("[!] No perforce support (ZOETK_P4_BIN)")
	} else {
		fmt.Println("ZOETK_P4_BIN:", analysis.P4_BIN)
	}
	if analysis.GIT_BIN == "" {
		fmt.Println("[!] No git support (ZOETK_GIT_BIN)")
	} else {
		fmt.Println("ZOETK_GIT_BIN:", analysis.GIT_BIN)
	}
	if analysis.CTAGS_BIN == "" {
		fmt.Println("[!] No ctags support (ZOETK_CTAGS_BIN)")
	} else {
		fmt.Println("ZOETK_CTAGS_BIN:", analysis.CTAGS_BIN)
	}

	if *version {
		fmt.Printf("zoekt-webserver version %q\n", Version)
		os.Exit(0)
	}

	if *logDir != "" {
		if fi, err := os.Lstat(*logDir); err != nil || !fi.IsDir() {
			log.Fatalf("%s is not a directory", *logDir)
		}
		// We could do fdup acrobatics to also redirect
		// stderr, but it is simpler and more portable for the
		// caller to divert stderr output if necessary.
		go divertLogs(*logDir, *logRefresh)
	}

	// Tune GOMAXPROCS to match Linux container CPU quota.
	maxprocs.Set()

	// TODO: init index searcher (*index -> path)

	s := &web.Server{
		// Searcher: searcher,
		Version:  Version,
		// IndexDir: *index,
	}

	s.SourceBaseDir = *fsbase
	s.HTML = *html
	s.BasicAuth.FileName = *basicauth

	if *hostCustomization != "" {
		s.HostCustomQueries = map[string]string{}
		for _, h := range strings.SplitN(*hostCustomization, ",", -1) {
			if len(h) == 0 {
				continue
			}
			fields := strings.SplitN(h, "=", 2)
			if len(fields) < 2 {
				log.Fatalf("invalid host_customization %q", h)
			}

			s.HostCustomQueries[fields[0]] = fields[1]
		}
	}

	handler, err := web.NewMux(s)
	if err != nil {
		log.Fatal(err)
	}

	handler.Handle("/metrics", promhttp.Handler())

	if *enablePprof {
		handler.HandleFunc("/debug/pprof/", pprof.Index)
		handler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		handler.HandleFunc("/debug/pprof/profile", pprof.Profile)
		handler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		handler.HandleFunc("/debug/pprof/trace", pprof.Trace)
		handler.HandleFunc("/debug/requests/", trace.Traces)
		handler.HandleFunc("/debug/events/", trace.Events)
	}

	watchdogAddr := "http://" + *listen
	if *sslCert != "" || *sslKey != "" {
		watchdogAddr = "https://" + *listen
	}
	go watchdog(30*time.Second, watchdogAddr)

	if *sslCert != "" || *sslKey != "" {
		log.Printf("serving HTTPS on %s", *listen)
		err = http.ListenAndServeTLS(*listen, *sslCert, *sslKey, handler)
	} else {
		log.Printf("serving HTTP on %s", *listen)
		err = http.ListenAndServe(*listen, handler)
	}
	log.Printf("ListenAndServe: %v", err)
}

func watchdogOnce(ctx context.Context, client *http.Client, addr string) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("watchdog: status %v", resp.StatusCode)
	}
	return nil
}

func watchdog(dt time.Duration, addr string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	tick := time.NewTicker(dt)

	errCount := 0
	for range tick.C {
		err := watchdogOnce(context.Background(), client, addr)
		if err != nil {
			errCount++
		} else {
			errCount = 0
		}
		if errCount == 3 {
			log.Panicf("watchdog: %v", err)
		}
	}
}
