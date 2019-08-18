package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/asfaltboy/urlshort"
	bolt "go.etcd.io/bbolt"
)

func main() {
	dbFile := flag.String("db", "example.db", "Path to bold db file (see https://godoc.org/go.etcd.io/bbolt)")
	yaml := flag.String("yaml", "example.yaml", "Path to yaml config file (see example.yaml)")
	json := flag.String("json", "", "Path to yaml config file (see example.yaml)")
	flag.Parse()

	if *yaml == "" && *json == "" && *dbFile == "" {
		log.Fatal("Must provide one source for path")
	}

	var mux http.Handler
	var err error
	mux = http.Handler(defaultMux())

	if *yaml != "" {
		mux, err = urlshort.YAMLHandler(readFileContent(*yaml), mux)
		if err != nil {
			log.Fatalf("cannot build yaml handler: %v", err)
		}
	}
	if *json != "" {
		mux, err = urlshort.JSONHandler(readFileContent(*json), mux)
		if err != nil {
			log.Fatalf("cannot build json handler: %v", err)
		}
	}
	if *dbFile != "" {
		db, err := bolt.Open(*dbFile, 0600, nil)
		if err != nil {
			log.Fatalf("error reading database file '%s': %v", *dbFile, err)
		}
		defer db.Close()
		mux, err = urlshort.BoltHandler(db, mux)
		if err != nil {
			log.Fatalf("cannot build bolt handler: %v", err)
		}
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", mux)
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

func readFileContent(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("could not open config file %s: %v", path, err)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("could not read config file: %v", err)
	}
	return content
}
