package traefik_plugin_redirect

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

// Config the plugin configuration.
type Config struct {
	Debug     bool     `json:"debug,omitempty"     yaml:"debug,omitempty"`
	Redirects []string `json:"redirects,omitempty" yaml:"redirects,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Plugin this is a Traefik redirect plugin.
type Plugin struct {
	next      http.Handler
	name      string
	debug     bool
	redirects []redirect
}

type redirect struct {
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	StatusCode  int    `json:"statusCode,omitempty"`
}

// New created a new plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	plugin := &Plugin{
		next:      next,
		name:      name,
		debug:     config.Debug,
		redirects: make([]redirect, 0),
	}

	for _, cfg := range config.Redirects {
		parts := strings.Split(cfg, ":")
		if len(parts) < 2 {
			continue
		}
		r := redirect{
			Source:      parts[0],
			Destination: parts[1],
			StatusCode:  302,
		}
		if len(parts) > 2 {
			if statusCode, err := strconv.Atoi(parts[2]); err == nil {
				r.StatusCode = statusCode
			}
		}

		plugin.redirects = append(plugin.redirects, r)
	}

	return plugin, nil
}

func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// If no redirection registered, skip to the next handler.
	if p.redirects == nil || len(p.redirects) == 0 {
		p.next.ServeHTTP(rw, req)
		return
	}

	// Loop all redirections
	for _, r := range p.redirects {
		if r.Source == req.RequestURI {
			// Add headers for debug
			if p.debug {
				rw.Header().Set("X-Middleware-Name", p.name)
				rw.Header().Set("X-Middleware-Source", r.Source)
				rw.Header().Set("X-Middleware-Destination", r.Destination)
				rw.Header().Set("X-Middleware-StatusCode", strconv.Itoa(r.StatusCode))
				rw.Header().Set("X-Middleware-Old-URL", req.RequestURI)
			}

			u, err := url.Parse(r.Destination)
			if err != nil {
				continue
			}

			// Check if identical url, and redirect
			handler := &moveHandler{location: u, statusCode: r.StatusCode}
			handler.ServeHTTP(rw, req)
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
