package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/*
var assetsFS embed.FS

func GenStaticFileServe(mux *http.ServeMux) {
       subFS, err := fs.Sub(assetsFS, "assets")
       if err != nil {
	       return
       }
       fileServer := http.FileServer(http.FS(subFS))
       mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))
}
