// Package localconsole serves the loopback console.
// The overview stays GET-only. The other pages accept a same-origin POST.
package localconsole

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/lucaswilliameufrasio/bornal/internal/console"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

// View is the data one overview request renders.
type View struct {
	Language i18n.Language
	Installs []install.Install
	Empty    string
	Running  string
}

// Form is one console POST. Passphrase is not written to the response or the log.
type Form struct {
	Kind       string
	Confirm    string
	Flavor     string
	Account    int
	Replace    bool
	RealmOK    bool
	Encrypt    bool
	Label      string
	Out        string
	Kit        string
	Passphrase []byte
	Again      []byte
}

// Actions serves the pages that can write. Overview stays GET-only.
type Actions struct {
	CSRF string
	Get  func(page, flavor string, account int, replace bool) (console.Page, error)
	Post func(page string, form Form) (console.Page, error)
}

// Server is a running console. URL is its overview page.
type Server struct {
	URL    string
	server *http.Server
}

// Start binds 127.0.0.1 on a port chosen by the operating system.
// current is called for the overview. actions serves the other pages.
func Start(current func() View, actions *Actions, report func(error)) (*Server, error) {
	if current == nil {
		return nil, errors.New("visão ausente")
	}
	if actions != nil && actions.CSRF == "" {
		actions.CSRF = newCSRF()
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	host := listener.Addr().String()
	server := &http.Server{
		Handler:           route(current, host, actions),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
		ErrorLog:          log.New(io.Discard, "", 0),
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

// newHandler answers the overview. Tests use it without the write pages.
func newHandler(current func() View, host string) http.Handler {
	return route(current, host, nil)
}

func route(current func() View, host string, actions *Actions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Cache-Control", "no-store")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		policy := "default-src 'none'; style-src 'unsafe-inline'; font-src data:; base-uri 'none'; frame-ancestors 'none'; form-action 'none'"
		if actions != nil && r.URL.Path != "/web/overview" && r.URL.Path != "/" && r.URL.Path != "/web" && r.URL.Path != "/web/" {
			policy = "default-src 'none'; style-src 'unsafe-inline'; font-src data:; base-uri 'none'; frame-ancestors 'none'; form-action 'self'"
		}
		header.Set("Content-Security-Policy", policy)

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
		if r.URL.Path == "/" || r.URL.Path == "/web" || r.URL.Path == "/web/" {
			if r.Method != http.MethodGet {
				header.Set("Allow", http.MethodGet)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			http.Redirect(w, r, "/web/overview", http.StatusFound)
			return
		}
		if r.URL.Path == "/web/overview" {
			if r.Method != http.MethodGet {
				header.Set("Allow", http.MethodGet)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			writePage(w, overview(current))
			return
		}
		pageName, ok := pageOf(r.URL.Path)
		if !ok || actions == nil || actions.Get == nil {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			account, _ := strconv.Atoi(r.URL.Query().Get("account"))
			page, err := actions.Get(pageName, r.URL.Query().Get("flavor"), account, r.URL.Query().Get("replace") == "1")
			if err != nil {
				http.Error(w, "page unavailable", http.StatusInternalServerError)
				return
			}
			writePage(w, page)
		case http.MethodPost:
			if r.URL.RawQuery != "" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			form, ok := readForm(w, r, actions.CSRF)
			if !ok {
				return
			}
			page, err := actions.Post(pageName, form)
			wipe(form.Passphrase)
			wipe(form.Again)
			if err != nil {
				http.Error(w, "page unavailable", http.StatusInternalServerError)
				return
			}
			writePage(w, page)
		default:
			header.Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func pageOf(path string) (string, bool) {
	switch path {
	case "/web/macros":
		return "macros", true
	case "/web/addons":
		return "addons", true
	case "/web/kits":
		return "kits", true
	case "/web/rollback":
		return "rollback", true
	default:
		return "", false
	}
}

func overview(current func() View) console.Page {
	view := current()
	return console.NewPage(view.Language, view.Installs, view.Empty, view.Running)
}

func writePage(w http.ResponseWriter, page console.Page) {
	var body bytes.Buffer
	if err := console.Render(&body, page); err != nil {
		http.Error(w, "page unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = body.WriteTo(w)
}

func readForm(w http.ResponseWriter, r *http.Request, csrf string) (Form, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return Form{}, false
	}
	got := r.PostForm.Get("csrf")
	if subtle.ConstantTimeCompare([]byte(got), []byte(csrf)) != 1 || csrf == "" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return Form{}, false
	}
	account, _ := strconv.Atoi(r.PostForm.Get("account"))
	form := Form{
		Kind:       r.PostForm.Get("kind"),
		Confirm:    r.PostForm.Get("confirm"),
		Flavor:     r.PostForm.Get("flavor"),
		Account:    account,
		Replace:    r.PostForm.Get("replace") == "1",
		RealmOK:    r.PostForm.Get("realm_ok") == "1",
		Encrypt:    r.PostForm.Get("encrypt") == "1",
		Label:      r.PostForm.Get("label"),
		Out:        r.PostForm.Get("out"),
		Kit:        r.PostForm.Get("kit"),
		Passphrase: []byte(r.PostForm.Get("passphrase")),
		Again:      []byte(r.PostForm.Get("passphrase2")),
	}
	r.PostForm.Del("passphrase")
	r.PostForm.Del("passphrase2")
	return form, true
}

func newCSRF() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(raw[:])
}

func wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
