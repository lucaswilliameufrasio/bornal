package localconsole

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucaswilliameufrasio/bornal/internal/console"
	"github.com/lucaswilliameufrasio/bornal/internal/i18n"
	"github.com/lucaswilliameufrasio/bornal/internal/install"
)

const testHost = "127.0.0.1:4100"

func view() View {
	return View{
		Language: i18n.Portuguese,
		Installs: []install.Install{{
			Path:     "/jogos/World of Warcraft/_forever_",
			Flavor:   "_forever_",
			Locale:   "enUS",
			AgentUID: "wowf",
		}},
	}
}

func serve(change func(*http.Request)) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/web/overview", nil)
	request.RemoteAddr = "127.0.0.1:50000"
	if change != nil {
		change(request)
	}
	recorder := httptest.NewRecorder()
	newHandler(view, testHost).ServeHTTP(recorder, request)
	return recorder
}

func TestOverviewRendersTheInstall(t *testing.T) {
	response := serve(nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status %d", response.Code)
	}
	body := response.Body.String()
	for _, want := range []string{"<html lang=\"pt-BR\">", "<h1>_forever_</h1>", "enUS", "wowf", "Para gravar", "data:font/woff2"} {
		if !strings.Contains(body, want) {
			t.Fatalf("ausente: %s", want)
		}
	}
	policy := response.Header().Get("Content-Security-Policy")
	if response.Header().Get("Cache-Control") != "no-store" || !strings.Contains(policy, "default-src 'none'") || !strings.Contains(policy, "font-src data:") {
		t.Fatal("cabeçalhos de proteção ausentes")
	}
}

func TestEnglishPage(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/web/overview", nil)
	request.RemoteAddr = "127.0.0.1:50000"
	recorder := httptest.NewRecorder()
	newHandler(func() View {
		return View{Language: i18n.English, Empty: "No installation at that path."}
	}, testHost).ServeHTTP(recorder, request)
	body := recorder.Body.String()
	if !strings.Contains(body, `lang="en"`) || !strings.Contains(body, "<h1>Installs</h1>") || !strings.Contains(body, "To write") {
		t.Fatalf("página: %s", body)
	}
}

func TestRefusesRequestsFromOutsideTheLocalBrowser(t *testing.T) {
	cases := map[string]func(*http.Request){
		"peer fora do loopback": func(r *http.Request) { r.RemoteAddr = "192.0.2.10:50000" },
		"host de outro nome":    func(r *http.Request) { r.Host = "exemplo.test:4100" },
		"origem de outro site":  func(r *http.Request) { r.Header.Set("Origin", "https://exemplo.test") },
		"fetch entre sites":     func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") },
	}
	for name, change := range cases {
		if response := serve(change); response.Code != http.StatusForbidden {
			t.Errorf("%s: status %d", name, response.Code)
		}
	}
}

func TestHandlerIsReadOnly(t *testing.T) {
	response := serve(func(r *http.Request) { r.Method = http.MethodPost })
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", response.Code)
	}
}

func TestRoutes(t *testing.T) {
	if response := serve(func(r *http.Request) { r.URL.Path = "/" }); response.Code != http.StatusFound || response.Header().Get("Location") != "/web/overview" {
		t.Fatalf("raiz: status %d para %q", response.Code, response.Header().Get("Location"))
	}
	if response := serve(func(r *http.Request) { r.URL.Path = "/web/outra" }); response.Code != http.StatusNotFound {
		t.Fatalf("caminho desconhecido: status %d", response.Code)
	}
}

func TestWritePagesKeepThePassphraseOffThePage(t *testing.T) {
	const secret = "s3cret-value"
	actions := &Actions{
		CSRF: "csrf-token",
		Get: func(page, flavor string, account int, replace bool) (console.Page, error) {
			out := console.Section(i18n.English, "Macros", "/web/macros", "note")
			out.Forms = []console.Form{{
				Kind: "apply", Action: "/web/macros", Submit: "Write macros", Password: true,
			}}
			return out, nil
		},
		Post: func(page string, form Form) (console.Page, error) {
			if page != "macros" || string(form.Passphrase) != secret || form.Kind != "apply" {
				t.Fatalf("form %+v", form.Kind)
			}
			out := console.Section(i18n.English, "Macros", "/web/macros", "note")
			out.Message = "wrote"
			return out, nil
		},
	}
	handler := route(view, testHost, actions)

	overview := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/web/overview", nil)
	overview.RemoteAddr = "127.0.0.1:50000"
	overviewRec := httptest.NewRecorder()
	handler.ServeHTTP(overviewRec, overview)
	if !strings.Contains(overviewRec.Header().Get("Content-Security-Policy"), "form-action 'none'") {
		t.Fatal("visão geral sem form-action none")
	}
	posted := httptest.NewRequest(http.MethodPost, "http://"+testHost+"/web/overview", strings.NewReader("csrf=csrf-token"))
	posted.RemoteAddr = "127.0.0.1:50000"
	posted.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postedRec := httptest.NewRecorder()
	handler.ServeHTTP(postedRec, posted)
	if postedRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("visão geral POST %d", postedRec.Code)
	}

	macros := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/web/macros", nil)
	macros.RemoteAddr = "127.0.0.1:50000"
	macrosRec := httptest.NewRecorder()
	handler.ServeHTTP(macrosRec, macros)
	policy := macrosRec.Header().Get("Content-Security-Policy")
	if !strings.Contains(policy, "form-action 'self'") || strings.Contains(policy, "form-action 'none'") {
		t.Fatalf("csp %s", policy)
	}

	query := httptest.NewRequest(http.MethodPost, "http://"+testHost+"/web/macros?x=1", strings.NewReader("csrf=csrf-token"))
	query.RemoteAddr = "127.0.0.1:50000"
	query.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	queryRec := httptest.NewRecorder()
	handler.ServeHTTP(queryRec, query)
	if queryRec.Code != http.StatusBadRequest {
		t.Fatalf("query %d", queryRec.Code)
	}

	bad := httptest.NewRequest(http.MethodPost, "http://"+testHost+"/web/macros", strings.NewReader("csrf=nope&passphrase="+secret))
	bad.RemoteAddr = "127.0.0.1:50000"
	bad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	badRec := httptest.NewRecorder()
	handler.ServeHTTP(badRec, bad)
	if badRec.Code != http.StatusForbidden || strings.Contains(badRec.Body.String(), secret) {
		t.Fatalf("csrf %d %s", badRec.Code, badRec.Body.String())
	}

	body := "csrf=csrf-token&kind=apply&passphrase=" + secret
	good := httptest.NewRequest(http.MethodPost, "http://"+testHost+"/web/macros", strings.NewReader(body))
	good.RemoteAddr = "127.0.0.1:50000"
	good.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	goodRec := httptest.NewRecorder()
	handler.ServeHTTP(goodRec, good)
	if goodRec.Code != http.StatusOK || !strings.Contains(goodRec.Body.String(), "wrote") || strings.Contains(goodRec.Body.String(), secret) {
		t.Fatalf("post %d %s", goodRec.Code, goodRec.Body.String())
	}
}

func TestPathIsEscaped(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/web/overview", nil)
	request.RemoteAddr = "127.0.0.1:50000"
	recorder := httptest.NewRecorder()
	newHandler(func() View {
		return View{
			Language: i18n.English,
			Installs: []install.Install{{Path: `<script>alert(1)</script>`, Flavor: "_forever_"}},
		}
	}, testHost).ServeHTTP(recorder, request)
	body := recorder.Body.String()
	if strings.Contains(body, "<script>alert") {
		t.Fatal("caminho entrou cru na página")
	}
}
