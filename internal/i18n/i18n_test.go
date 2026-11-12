package i18n

import "testing"

func TestDetect(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want Language
	}{
		{name: "vazio", env: map[string]string{}, want: English},
		{name: "inglês", env: map[string]string{"LANG": "en_US.UTF-8"}, want: English},
		{name: "brasil", env: map[string]string{"LANG": "pt_BR.UTF-8"}, want: Portuguese},
		{name: "portugal cai no texto brasileiro", env: map[string]string{"LANG": "pt_PT.UTF-8"}, want: Portuguese},
		{name: "LC_ALL pt ganha de LANG en", env: map[string]string{"LANG": "en_US.UTF-8", "LC_ALL": "pt_BR.UTF-8"}, want: Portuguese},
		{name: "LANG pt ganha de LC_ALL en", env: map[string]string{"LANG": "pt_BR.UTF-8", "LC_ALL": "en_US.UTF-8"}, want: Portuguese},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Detect(func(key string) string { return tc.env[key] })
			if got != tc.want {
				t.Fatalf("Detect = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	lang, ok := Parse("pt-BR")
	if !ok || lang != Portuguese {
		t.Fatalf("pt-BR: %v %v", lang, ok)
	}
	lang, ok = Parse("en")
	if !ok || lang != English {
		t.Fatalf("en: %v %v", lang, ok)
	}
	if _, ok = Parse("de"); ok {
		t.Fatal("de foi aceito")
	}
}
