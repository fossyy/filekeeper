package session

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func cookieByName(t *testing.T, writer *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range writer.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not set", name)
	return nil
}

func TestSessionCookieFlags(t *testing.T) {
	t.Setenv("HOSTNAME", "test")
	t.Setenv("SESSION_NAME", "Session")
	t.Setenv("SESSION_MAX_AGE", "3600")

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	tlsRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	tlsRequest.TLS = &tls.ConnectionState{}
	proxiedRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	proxiedRequest.Header.Set("X-Forwarded-Proto", "https")

	tests := []struct {
		name    string
		request *http.Request
	}{
		{"plain http", request},
		{"tls request", tlsRequest},
		{"proxied https", proxiedRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := httptest.NewRecorder()
			(&Session{ID: "id"}).Save(test.request, writer)
			cookie := cookieByName(t, writer, "Session")

			if cookie.Value != "id" {
				t.Errorf("cookie value = %q, want %q", cookie.Value, "id")
			}
			if cookie.MaxAge != 3600 {
				t.Errorf("cookie MaxAge = %d, want 3600", cookie.MaxAge)
			}
			if cookie.Path != "/" {
				t.Errorf("cookie Path = %q, want %q", cookie.Path, "/")
			}
			if !cookie.HttpOnly {
				t.Error("cookie HttpOnly = false, want true")
			}
			if cookie.Secure != IsHTTPS(test.request) {
				t.Errorf("cookie Secure = %v, want %v", cookie.Secure, IsHTTPS(test.request))
			}
			if cookie.SameSite != http.SameSiteLaxMode {
				t.Errorf("cookie SameSite = %v, want Lax", cookie.SameSite)
			}
		})
	}
}

func TestDestroyCookieClear(t *testing.T) {
	t.Setenv("HOSTNAME", "test")
	t.Setenv("SESSION_NAME", "Session")
	t.Setenv("SESSION_MAX_AGE", "3600")

	writer := httptest.NewRecorder()
	(&Session{ID: "id"}).Destroy(httptest.NewRequest(http.MethodGet, "/", nil), writer)
	cookie := cookieByName(t, writer, "Session")

	if cookie.Value != "" {
		t.Errorf("cookie value = %q, want empty", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("cookie MaxAge = %d, want -1", cookie.MaxAge)
	}
	if cookie.Path != "/" {
		t.Errorf("cookie Path = %q, want %q", cookie.Path, "/")
	}
	if !cookie.HttpOnly {
		t.Error("cookie HttpOnly = false, want true")
	}
	if cookie.Secure {
		t.Error("cookie Secure = true, want false over plain http")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite = %v, want Lax", cookie.SameSite)
	}
}
