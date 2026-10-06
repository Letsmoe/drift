package main

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

var errBackendStarting = errors.New("the backend is still starting")

// newServerHandler routes the backend's single loopback origin: /stream and
// /cover go to the proxy, /rpc to the hub, everything else to the SPA.
func newServerHandler(proxy http.Handler, rpc http.Handler, assets fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/stream/", proxy)
	mux.Handle("/cover/", proxy)
	mux.Handle("/rpc", rpc)
	mux.Handle("/", spaHandler(assets))
	return mux
}

// spaHandler serves the static SvelteKit build and falls back to index.html for
// client-side routes, like adapter-static's fallback expects.
func spaHandler(assets fs.FS) http.Handler {
	files := http.FileServerFS(assets)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		name := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if name == "" || !fileExists(assets, name) {
			http.ServeFileFS(writer, request, assets, "index.html")
			return
		}
		files.ServeHTTP(writer, request)
	})
}

func fileExists(assets fs.FS, name string) bool {
	info, err := fs.Stat(assets, name)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
