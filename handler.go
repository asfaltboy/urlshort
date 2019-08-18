package urlshort

import (
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v2"
)

type mapPathHandler struct {
	pathMap  map[string]string
	fallback http.Handler
}

// Generate a path map from a list of maps as per marshalled config, and a fallback
func newMapPathHandler(shortPaths []marshalledConfig, fallback http.Handler) mapPathHandler {
	// Note: duplicate URLs are squashed
	pathsToUrls := map[string]string{}
	for _, v := range shortPaths {
		pathsToUrls[v.Path] = v.URL
	}
	return mapPathHandler{pathsToUrls, fallback}
}

func (m *mapPathHandler) redirectToPath(w http.ResponseWriter, r *http.Request) {
	url := m.pathMap[r.URL.Path]
	if url != "" {
		http.Redirect(w, r, url, http.StatusFound)
	} else {
		m.fallback.ServeHTTP(w, r)
	}
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	m := mapPathHandler{pathsToUrls, fallback}
	return http.HandlerFunc(m.redirectToPath)
}

type marshalledConfig struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc that will attempt to map any paths to their
// corresponding URL. If the path is not provided in the YAML,
// then the fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned are related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	// parse YAML input
	yamlPaths := []marshalledConfig{}
	err := yaml.UnmarshalStrict(yml, &yamlPaths)
	if err != nil {
		return nil, err
	}
	m := newMapPathHandler(yamlPaths, fallback)
	return http.HandlerFunc(m.redirectToPath), err
}

// JSONHandler will parse the privided JSON and return
// an http.HandlerFunc that will attempt to map any paths to their
// corresponding URL. If the path is not provided in the YAML,
// then the fallback http.Handler will be called instead.
//
// JSON is expected to be in the format:
//
//     { "path": "/some-path",
//       "url": "https://www.some-url.com/demo" }
//
// The only errors that can be returned are related to having
// invalid JSON data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func JSONHandler(j []byte, fallback http.Handler) (http.HandlerFunc, error) {
	// parse JSON input
	jsonPaths := []marshalledConfig{}
	err := json.Unmarshal(j, &jsonPaths)
	if err != nil {
		return nil, err
	}
	m := newMapPathHandler(jsonPaths, fallback)
	return http.HandlerFunc(m.redirectToPath), err
}
