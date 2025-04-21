package main

import "net/http"

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-10")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))

}