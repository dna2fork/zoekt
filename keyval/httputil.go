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

func ServeBasic(w http.ResponseWriter, r *http.Request) {
	// /keyval/_?k=keyval://path/to/sth&sk=index/0000
	// -> /storage/keyval/path/to/sth/_/index/0000

	// k=, sk=, a=, <body>
	// a=list -> k.<sub keys>
	// POST:<body> -> put
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
	all := false
	if subkey == "" {
		subkey = "_"
		all = true
	}
	key0 := defaultStorage.WithSubkey(key, subkey)
	_, ok := defaultStorage.Get(key0)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if all {
		list := defaultStorage.ListSubkey(key)
		for _, sk := range list {
			ok = defaultStorage.Del(defaultStorage.WithSubkey(key, sk))
			// TODO: !ok
		}
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(url + "#" + strings.Join(list, ",")))
		return
	}
	ok = defaultStorage.Del(key0)
	if !ok {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%s#%s", url, subkey)))
}

