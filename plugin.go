package traefik_plugin_redirect

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

// Config the plugin configuration.
type Config struct {
	Regex       string `json:"regex"`
	Replacement string `json:"replacement"`
	StatusCode  int    `json:"statusCode"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Regex:       "",
		Replacement: "",
		StatusCode:  0,
	}
}

// Redirect a plugin.
type Redirect struct {
	next        http.Handler
	name        string
	regex       *regexp.Regexp
	replacement string
	statusCode  int
	rawURL      func(*http.Request) string
}

// New created a new plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rxp, err := regexp.Compile(config.Regex)
	if err != nil {
		return nil, err
	}

	return &Redirect{
		next:        next,
		name:        name,
		regex:       rxp,
		replacement: config.Replacement,
		statusCode:  config.StatusCode,
		rawURL:      rawURL,
	}, nil
}

func (r *Redirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	oldURL := r.rawURL(req)

	// If the Regexp doesn't match, skip to the next handler.
	if !r.regex.MatchString(oldURL) {
		r.next.ServeHTTP(rw, req)
		return
	}

	// Apply a rewrite regexp to the URL.
	newURL := r.regex.ReplaceAllString(oldURL, r.replacement)

	// Parse the rewritten URL and replace request URL with it.
	parsedURL, err := url.Parse(newURL)
	if err != nil {
		r.next.ServeHTTP(rw, req)
		return
	}

	if newURL != oldURL {
		handler := &moveHandler{location: parsedURL, statusCode: r.statusCode}
		handler.ServeHTTP(rw, req)
		return
	}

	req.URL = parsedURL

	// Make sure the request URI corresponds the rewritten URL.
	req.RequestURI = req.URL.RequestURI()
	r.next.ServeHTTP(rw, req)
}

type moveHandler struct {
	location   *url.URL
	statusCode int
}

func (m *moveHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Location", m.location.String())

	status := m.statusCode
	if status == 0 {
		status = http.StatusFound
		if req.Method != http.MethodGet {
			status = http.StatusTemporaryRedirect
		}
	}

	if req.Method != http.MethodGet && status == http.StatusMovedPermanently {
		status = http.StatusPermanentRedirect
	}

	rw.WriteHeader(status)
	_, err := rw.Write([]byte(http.StatusText(status)))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func rawURL(req *http.Request) string {
	scheme := schemeHTTP
	host := req.Host
	port := ""
	uri := req.RequestURI

	schemeRegex := `^(https?):\/\/(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`
	re, _ := regexp.Compile(schemeRegex)
	if re.Match([]byte(req.RequestURI)) {
		match := re.FindStringSubmatch(req.RequestURI)
		scheme = match[1]

		if len(match[2]) > 0 {
			host = match[2]
		}

		if len(match[3]) > 0 {
			port = match[3]
		}

		uri = match[4]
	}

	if req.TLS != nil {
		scheme = schemeHTTPS
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}
