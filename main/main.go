package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/asfaltboy/urlshort"
)

func main() {
	path := flag.String("config", "example.yaml", "Path to yaml config file (see example.yaml)")
	flag.Parse()
	file, err := os.Open(*path)
	if err != nil {
		log.Fatalf("could not open config file %s: %v", *path, err)
	}
	yaml, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("could not parse config file: %v", err)
	}

	// Build the YAMLHandler using default 404 fallback mux
	fallback := defaultMux()
	yamlHandler, err := urlshort.YAMLHandler([]byte(yaml), fallback)
	if err != nil {
		log.Panicf("cannot build yaml handler: %v", err)
	}
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", yamlHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", notFound)
	return mux
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "Unknown urlshort entry!")
}
