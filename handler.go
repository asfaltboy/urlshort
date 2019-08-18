package urlshort

import (
	"encoding/json"
	"errors"
	"net/http"

	bolt "go.etcd.io/bbolt"
	"gopkg.in/yaml.v2"
)

type mapPathHandler struct {
	pathMap  map[string]string
	fallback http.Handler
}

type pathConfig struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}

// Generate a path map from a list of maps as per marshalled config, and a fallback
func pathConfigToHandler(shortPaths []pathConfig, fallback http.Handler) mapPathHandler {
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
	yamlPaths := []pathConfig{}
	err := yaml.UnmarshalStrict(yml, &yamlPaths)
	if err != nil {
		return nil, err
	}
	m := pathConfigToHandler(yamlPaths, fallback)
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
	jsonPaths := []pathConfig{}
	err := json.Unmarshal(j, &jsonPaths)
	if err != nil {
		return nil, err
	}
	m := pathConfigToHandler(jsonPaths, fallback)
	return http.HandlerFunc(m.redirectToPath), nil
}

// BoltHandler will load url paths from the given DB instance
// and return http.HandlerFunc. If the path is not found
// in the db, then the fallback http.Handler will be called
// instead.
func BoltHandler(db *bolt.DB, fallback http.Handler) (http.HandlerFunc, error) {
	paths := map[string]string{}
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urlshort"))
		if b == nil {
			return errors.New("Db missing bucket 'urlshort'")
		}
		if err := b.ForEach(func(key, value []byte) error {
			paths[string(key)] = string(value)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	m := mapPathHandler{paths, fallback}
	return http.HandlerFunc(m.redirectToPath), nil
}
