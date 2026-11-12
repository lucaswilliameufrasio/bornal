package localconsole

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	for _, want := range []string{"<html lang=\"pt-BR\">", "<h1>Instalações</h1>", "_forever_", "enUS", "wowf"} {
		if !strings.Contains(body, want) {
			t.Fatalf("ausente: %s", want)
		}
	}
	if response.Header().Get("Cache-Control") != "no-store" || !strings.Contains(response.Header().Get("Content-Security-Policy"), "default-src 'none'") {
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
	if !strings.Contains(body, `lang="en"`) || !strings.Contains(body, "<h1>Installs</h1>") {
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
