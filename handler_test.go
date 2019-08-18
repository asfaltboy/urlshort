package urlshort

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Build the fallback default mux
func getDefaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})
	return mux
}

func getMapHandler(fallback *http.ServeMux) http.Handler {
	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}

	return MapHandler(pathsToUrls, fallback)
}

func TestMapHandlerMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/urlshort-godoc", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := getMapHandler(nil)
	rr := httptest.NewRecorder()

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

	mux := getDefaultMux()
	handler := getMapHandler(mux)
	rr := httptest.NewRecorder()
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

// Build the JSONHandler using given fallback
func getYAMLHandler(fallback *http.ServeMux) http.Handler {
	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution`

	handler, err := YAMLHandler([]byte(yaml), fallback)
	if err != nil {
		panic(err)
	}
	return handler
}

func TestYAMLHandlerMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/urlshort", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Build the YAMLHandler using given fallback
	handler := getYAMLHandler(nil)
	rr := httptest.NewRecorder()
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
	mux := getDefaultMux()
	handler := getYAMLHandler(mux)
	rr := httptest.NewRecorder()
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

// Build the JSONHandler using given fallback
func getJSONHandler(fallback *http.ServeMux) http.Handler {
	json := `[{"path": "/urlshort", "url": "https://github.com/gophercises/urlshort"},
{"path": "/urlshort-final", "url": "https://github.com/gophercises/urlshort/tree/solution"}]`

	handler, err := JSONHandler([]byte(json), fallback)
	if err != nil {
		panic(err)
	}
	return handler
}

func TestJSONHandlerMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/urlshort", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := getYAMLHandler(nil)

	rr := httptest.NewRecorder()
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

func TestJSONHandlerNotMatched(t *testing.T) {
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Build the JSONHandler using the mux as the fallback
	mux := getDefaultMux()
	handler := getJSONHandler(mux)

	rr := httptest.NewRecorder()
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

func TestInvalidJSONHandler(t *testing.T) {
	yaml := `foo`

	_, err := JSONHandler([]byte(yaml), nil)
	if err == nil {
		t.Errorf("Handler did not return error")
	}
}
