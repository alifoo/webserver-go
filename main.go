package main

import (
	"fmt"
	"net/http"
	"os"
)

func ReadinessEndpoint(h http.ResponseWriter, req *http.Request) {
	h.Header().Set("Content-Type", "text/plain; charset=utf-8")
	h.WriteHeader(http.StatusOK)

	h.Write([]byte("OK"))

}

func main() {
	mux := http.NewServeMux()
	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/healthz", ReadinessEndpoint)
	staticPath := http.Dir("./")
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(staticPath)))

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}