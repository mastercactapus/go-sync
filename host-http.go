package main

import (
	"log"
	"net/http"
)

func HostHTTP(path string, addr string) {
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(path))))
}
