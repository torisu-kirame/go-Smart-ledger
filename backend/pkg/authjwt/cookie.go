package authjwt

import (
	"net/http"
	"time"
)

const RefreshCookieName = "sl_refresh_token"

// RefreshCookieOptions holds secure cookie settings (mitigate XSS / CSRF).
type RefreshCookieOptions struct {
	Secure   bool
	Domain   string
	MaxAge   int
	SameSite http.SameSite
}

func DefaultRefreshCookieOpts(refreshExpireSec int, secure bool) RefreshCookieOptions {
	return RefreshCookieOptions{
		Secure:   secure,
		MaxAge:   refreshExpireSec,
		SameSite: http.SameSiteStrictMode,
	}
}

func SetRefreshCookie(w http.ResponseWriter, token string, opt RefreshCookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		MaxAge:   opt.MaxAge,
		HttpOnly: true,
		Secure:   opt.Secure,
		SameSite: opt.SameSite,
		Domain:   opt.Domain,
	})
}

func ClearRefreshCookie(w http.ResponseWriter, opt RefreshCookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   opt.Secure,
		SameSite: opt.SameSite,
		Domain:   opt.Domain,
	})
}

func ReadRefreshCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return "", err
	}
	if c.Value == "" {
		return "", http.ErrNoCookie
	}
	return c.Value, nil
}

// AccessFromHeader extracts Bearer access token.
func AccessFromHeader(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && h[:7] == "Bearer " {
		return h[7:]
	}
	return ""
}

// WithCookieMaxAge helper for refresh handler timing.
func (o RefreshCookieOptions) ExpiresAt() time.Time {
	return time.Now().Add(time.Duration(o.MaxAge) * time.Second)
}
