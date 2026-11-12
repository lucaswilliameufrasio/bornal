// Package localconsole serves the read-only overview on IPv4 loopback.
package localconsole

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/lucaswilliameufrasio/bornal/internal/console"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

// View is the data one request renders.
type View struct {
	Language i18n.Language
	Installs []install.Install
	Empty    string
}

// Server is a running console. URL is its overview page.
type Server struct {
	URL    string
	server *http.Server
}

// Start binds 127.0.0.1 on a port chosen by the operating system.
// current is called for every page so a later slice can reread the install.
func Start(current func() View, report func(error)) (*Server, error) {
	if current == nil {
		return nil, errors.New("visão ausente")
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	host := listener.Addr().String()
	server := &http.Server{
		Handler:           newHandler(current, host),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) && report != nil {
			report(err)
		}
	}()
	return &Server{URL: "http://" + host + "/web/overview", server: server}, nil
}

// Close stops the listener and drops open connections.
func (s *Server) Close() error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Close()
}

// newHandler answers only GET from a loopback peer addressed to host.
func newHandler(current func() View, host string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Cache-Control", "no-store")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")

		peerHost, _, err := net.SplitHostPort(r.RemoteAddr)
		peer := net.ParseIP(peerHost)
		origin := r.Header.Get("Origin")
		fetchSite := r.Header.Get("Sec-Fetch-Site")
		switch {
		case err != nil || peer == nil || !peer.IsLoopback(),
			r.Host != host,
			origin != "" && origin != "http://"+host,
			fetchSite != "" && fetchSite != "none" && fetchSite != "same-origin":
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet {
			header.Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/", "/web", "/web/":
			http.Redirect(w, r, "/web/overview", http.StatusFound)
			return
		case "/web/overview":
		default:
			http.NotFound(w, r)
			return
		}

		view := current()
		var body bytes.Buffer
		if err := console.Render(&body, console.NewPage(view.Language, view.Installs, view.Empty)); err != nil {
			http.Error(w, "page unavailable", http.StatusInternalServerError)
			return
		}
		header.Set("Content-Type", "text/html; charset=utf-8")
		_, _ = body.WriteTo(w)
	})
}
