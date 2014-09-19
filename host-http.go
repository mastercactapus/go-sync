package main

import (
	"log"
	"net/http"
	"strconv"
)

func HostHTTP(path string, port uint16) {
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(port)), http.FileServer(http.Dir(path))))
}
