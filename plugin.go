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

// Redirect holds one redirect configuration
type Redirect struct {
	Regex       string `json:"regex" yaml:"regex"`
	Replacement string `json:"replacement" yaml:"replacement"`
	StatusCode  int    `json:"statusCode" yaml:"statusCode"`
}

// Config the plugin configuration.
type Config struct {
	Redirects []Redirect `json:"redirects,omitempty" yaml:"redirects,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Plugin this is a Traefik redirect plugin.
type Plugin struct {
	next      http.Handler
	name      string
	redirects []redirect
	rawURL    func(*http.Request) string
}

type redirect struct {
	Regex       *regexp.Regexp `json:"regex,omitempty"`
	Replacement string         `json:"replacement,omitempty"`
	StatusCode  int            `json:"statusCode,omitempty"`
}

// New created a new plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	plugin := &Plugin{
		next:      next,
		name:      name,
		redirects: make([]redirect, 0),
		rawURL:    rawURL,
	}

	for _, cfg := range config.Redirects {
		rxp, err := regexp.Compile(cfg.Regex)
		if err != nil {
			return nil, err
		}

		plugin.redirects = append(plugin.redirects, redirect{
			Regex:       rxp,
			Replacement: cfg.Replacement,
			StatusCode:  cfg.StatusCode,
		})
	}

	return plugin, nil
}

func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	oldURL := p.rawURL(req)

	// If no redirection registered, skip to the next handler.
	if p.redirects == nil || len(p.redirects) == 0 {
		p.next.ServeHTTP(rw, req)
		return
	}

	// Loop all redirections
	for _, r := range p.redirects {
		if r.Regex.MatchString(oldURL) {
			// Apply a rewrite regexp to the URL.
			newURL := r.Regex.ReplaceAllString(oldURL, r.Replacement)

			// Parse the rewritten URL and replace request URL with it.
			parsedURL, err := url.Parse(newURL)
			if err != nil {
				continue
			}

			// Check if identical url, and redirect
			if newURL != oldURL {
				handler := &moveHandler{location: parsedURL, statusCode: r.StatusCode}
				handler.ServeHTTP(rw, req)
				return
			}

			// Make sure the request URI corresponds the rewritten URL.
			req.URL = parsedURL
			req.RequestURI = req.URL.RequestURI()
			p.next.ServeHTTP(rw, req)
			return
		}
	}

	p.next.ServeHTTP(rw, req)
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
