package middleware

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hgyowan/go-pkg-library/envs"
	"golang.org/x/net/publicsuffix"
	"google.golang.org/grpc/metadata"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func SessionCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = r.Header.Get("Referer")
		}

		if origin == "" {
			http.Error(w, "Forbidden: missing origin", http.StatusForbidden)
			return
		}

		u, err := url.Parse(origin)
		if err != nil {
			http.Error(w, "Forbidden: invalid origin", http.StatusForbidden)
			return
		}

		domain := u.Host
		if !strings.Contains(envs.CFMCookieDomain, domain) {
			http.Error(w, "Forbidden: unauthorized domain", http.StatusForbidden)
			return
		}

		secure := false
		host := u.Hostname()
		if envs.ServiceType != "local" {
			secure = true

			eTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(host)
			if err != nil {
				http.Error(w, "Forbidden: unauthorized tld domain", http.StatusForbidden)
				return
			}

			host = eTLDPlusOne
		}

		sid := uuid.NewString()
		cookie := &http.Cookie{
			Name:     "sid",
			Value:    sid,
			Path:     "/",
			Domain:   fmt.Sprintf(".%s", host),
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(24 * time.Hour * 7),
		}
		http.SetCookie(w, cookie)

		newCtx := metadata.AppendToOutgoingContext(r.Context(), "sessionID", sid)
		r = r.WithContext(newCtx)

		next.ServeHTTP(w, r)
	})
}

func GetSessionCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("sid")
		if err != nil {
			http.Error(w, "Forbidden: session id is required", http.StatusForbidden)
			return
		}
		newCtx := metadata.AppendToOutgoingContext(r.Context(), "sessionID", c.Value)
		r = r.WithContext(newCtx)

		next.ServeHTTP(w, r)
	})
}
