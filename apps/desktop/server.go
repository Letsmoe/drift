package main

import (
	"errors"
	"net/http"
)

var errBackendStarting = errors.New("the backend is still starting")

// newServerHandler routes the backend's single loopback origin: /stream and
// /cover go to the proxy, /rpc to the hub and the Neoworks sign-in callback to
// neoworksCallback. The UI is not served from here: the shell runs it.
func newServerHandler(proxy, rpc, neoworksCallback http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/stream/", proxy)
	mux.Handle("/cover/", proxy)
	mux.Handle("/rpc", rpc)
	mux.Handle(neoworksCallbackPath, neoworksCallback)
	return mux
}
