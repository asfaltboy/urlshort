package urlshort

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMapHandlerMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/urlshort-godoc", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}

	rr := httptest.NewRecorder()
	handler := MapHandler(pathsToUrls, nil)

	handler.ServeHTTP(rr, req)

	res := rr.Result()
	if status := res.StatusCode; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}

	expected := "https://godoc.org/github.com/gophercises/urlshort"
	fmt.Println(res.Header)
	if expected != res.Header.Get("Location") {
		t.Errorf("Handler returned wrong location: got %v want %v", expected, res.Header.Get("Location"))
	}
}

func TestMapHandlerNotMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{}

	rr := httptest.NewRecorder()
	handler := MapHandler(pathsToUrls, mux)

	handler.ServeHTTP(rr, req)

	res := rr.Result()
	if status := res.StatusCode; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Hello, world!"
	actual, _ := ioutil.ReadAll(res.Body)
	if expected != string(actual) {
		t.Errorf("Handler returned wrong body: got %v want %v", string(actual), expected)
	}
}

func TestYAMLHandlerMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/urlshort", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Build the YAMLHandler using the mux as the fallback
	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`

	rr := httptest.NewRecorder()
	handler, err := YAMLHandler([]byte(yaml), nil)
	if err != nil {
		panic(err)
	}

	handler.ServeHTTP(rr, req)

	res := rr.Result()
	if status := res.StatusCode; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}

	expected := "https://github.com/gophercises/urlshort"
	fmt.Println(res.Header)
	if expected != res.Header.Get("Location") {
		t.Errorf("Handler returned wrong location: got %v want %v", res.Header.Get("Location"), expected)
	}
}

func TestYAMLHandlerNotMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Build the YAMLHandler using the mux as the fallback
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})

	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
`

	rr := httptest.NewRecorder()
	handler, err := YAMLHandler([]byte(yaml), mux)
	if err != nil {
		panic(err)
	}

	handler.ServeHTTP(rr, req)

	res := rr.Result()
	if status := res.StatusCode; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Hello, world!"
	actual, _ := ioutil.ReadAll(res.Body)
	if expected != string(actual) {
		t.Errorf("Handler returned wrong location: got %v want %v", actual, expected)
	}
}

func TestInvalidYAMLHandler(t *testing.T) {
	yaml := `foo`

	_, err := YAMLHandler([]byte(yaml), nil)
	if err == nil {
		t.Errorf("Handler did not return error")
	}
}
