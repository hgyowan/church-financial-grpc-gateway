package middleware

import (
	"github.com/google/uuid"
	"github.com/hgyowan/go-pkg-library/envs"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

func SessionCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secure := true
		if envs.ServiceType == "local" {
			secure = false
		}

		sid := uuid.NewString()
		cookie := &http.Cookie{
			Name:     "sid",
			Value:    sid,
			Path:     "/",
			Domain:   envs.CFMCookieDomain,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(24 * time.Hour * 7),
		}
		http.SetCookie(w, cookie)

		newCtx := metadata.AppendToOutgoingContext(r.Context(), "sessionID", sid)
		r = r.WithContext(newCtx)

		// 다음 핸들러 호출
		next.ServeHTTP(w, r)
	})
}

func GetSessionCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Cookie("sid")
		newCtx := metadata.AppendToOutgoingContext(r.Context(), "sessionID", c.Value)
		r = r.WithContext(newCtx)

		// 다음 핸들러 호출
		next.ServeHTTP(w, r)
	})
}
