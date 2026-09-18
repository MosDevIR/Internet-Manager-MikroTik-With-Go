package i18n

import (
	"net/http"
	"strings"
)

const CookieName = "lang"

type Lang string

const (
	FA Lang = "fa"
	EN Lang = "en"
)

func Parse(s string) Lang {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(s, "en") {
		return EN
	}
	return FA
}

func FromRequest(r *http.Request) Lang {
	if q := r.URL.Query().Get("lang"); q != "" {
		return Parse(q)
	}
	if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
		return Parse(c.Value)
	}
	if al := r.Header.Get("Accept-Language"); strings.HasPrefix(strings.ToLower(al), "en") {
		return EN
	}
	return FA
}

func SetCookie(w http.ResponseWriter, lang Lang) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    string(lang),
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func T(lang Lang, key string) string {
	if lang == EN {
		if v, ok := en[key]; ok {
			return v
		}
	}
	if v, ok := fa[key]; ok {
		return v
	}
	if v, ok := en[key]; ok {
		return v
	}
	return key
}

// Dict returns a map of all keys for the given language (for templates).
func Dict(lang Lang) map[string]string {
	src := fa
	if lang == EN {
		src = en
	}
	out := make(map[string]string, len(src)+len(en))
	for k, v := range fa {
		out[k] = v
	}
	if lang == EN {
		for k, v := range en {
			out[k] = v
		}
	}
	return out
}

func IsRTL(lang Lang) bool {
	return lang != EN
}

func HTMLLang(lang Lang) string {
	if lang == EN {
		return "en"
	}
	return "fa"
}

func Dir(lang Lang) string {
	if lang == EN {
		return "ltr"
	}
	return "rtl"
}
