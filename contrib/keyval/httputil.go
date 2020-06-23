package keyval

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
)

const (
	maxValueSize = 1024 * 1024 * 4
)

var defaultStorage = StorageFilesystem{}

func StorageHandler(w http.ResponseWriter, r *http.Request) {
	// modified original handler to bind to a specific path
	qv := r.URL.Query()
	cmd := qv.Get("a")
	switch cmd {
	default:
		serveBasic(w, r)
	}
}

func serveBasic(w http.ResponseWriter, r *http.Request) {
	// /keyval/_?k=keyval://path/to/sth&sk=index/0000
	// -> /storage/keyval/path/to/sth/_/index/0000
	qv := r.URL.Query()
	key := qv.Get("k")
	if key == "" {
		http.NotFound(w, r)
		return
	}
	if !strings.Contains(key, "://") {
		// must follow protocol://path/to/sth...
		http.NotFound(w, r)
		return
	}
	subkey := qv.Get("sk")
	switch r.Method {
	case "GET":
		getValue(key, subkey, w, r)
		return
	case "POST":
		putKeyValue(key, subkey, w, r)
		return
	case "PUT":
		putKeyValue(key, subkey, w, r)
		return
	case "DELETE":
		delKey(key, subkey, w, r)
		return
	}
	// PATCH, OPTION, ...
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprint(w, "")
}

func getValue(url string, subkey string, w http.ResponseWriter, r *http.Request) {
	// e.g. ?k=keyval://test               --> /storage/keyval/test/_/_
	//      ?k=keyval://test&sk=index/0000 --> /storage/keyval/test/_/index/0000
	//      ?k=keyval://test&a=list        --> _
	//                                         index/
	qv := r.URL.Query()
	a := qv.Get("a")
	key := defaultStorage.UrlToKey(url)
	switch a {
	case "list":
		if subkey != "" {
			key = defaultStorage.WithSubkey(key, subkey)
		}
		getKeySubkeyList(key, w, r)
	default:
		if subkey == "" {
			subkey = "_"
		}
		key = defaultStorage.WithSubkey(key, subkey)
		getKeyValue(key, w, r)
	}
}

func getKeyValue(key string, w http.ResponseWriter, r *http.Request) {
	b, ok := defaultStorage.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.Write(b)
}

func getKeySubkeyList(key string, w http.ResponseWriter, r *http.Request) {
	list := defaultStorage.ListSubkey(key)
	b := []byte(strings.Join(list, "\n"))
	w.Header().Add("Content-Type", "text/plain")
	w.Write(b)
}

func putKeyValue(url string, subkey string, w http.ResponseWriter, r *http.Request) {
	key := defaultStorage.UrlToKey(url)
	if subkey == "" {
		subkey = "_"
	}
	key = defaultStorage.WithSubkey(key, subkey)
	r.Body = http.MaxBytesReader(w, r.Body, maxValueSize)	
	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	ok := defaultStorage.Put(key, value)
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%s#%s", url, subkey)))
}

func delKey(url string, subkey string, w http.ResponseWriter, r *http.Request) {
	key := defaultStorage.UrlToKey(url)
	if subkey == "" {
		subkey = "_"
	}
	key = defaultStorage.WithSubkey(key, subkey)
	_, ok := defaultStorage.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	ok = defaultStorage.Del(key)
	if !ok {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%s#%s", url, subkey)))
}

