package middleware

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	accountModel "github.com/hgyowan/church-financial-account-grpc/domain/token"
	"github.com/hgyowan/go-pkg-library/envs"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strconv"
	"strings"
)

func ValidTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(envs.UserTokenHeaderName)
		if token == "" {
			token = r.URL.Query().Get(envs.UserTokenHeaderName)
			if token == "" {
				http.Error(w, "token is required", http.StatusUnauthorized)
				return
			}
		}

		var claims = &accountModel.JWTCustomClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(envs.JwtAccessSecret), nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				http.Error(w, "token is expired", http.StatusUnauthorized)
				return
			}

			http.Error(w, "Authentication failed", http.StatusForbidden)
			return
		}

		// 메타데이터를 컨텍스트에 추가
		newCtx := metadata.AppendToOutgoingContext(r.Context(), "user_id", strconv.Itoa(int(claims.UserID)), "ip", getClientIP(r), "access_token", token)
		r = r.WithContext(newCtx)

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	// 1. X-Forwarded-For 헤더 확인
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For 값이 여러 개일 수 있으므로 첫 번째 값을 가져옴
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 2. X-Real-IP 헤더 확인
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 3. 기본 RemoteAddr 사용
	return r.RemoteAddr
}
