package middleware

import (
	"context"
	"google.golang.org/grpc/metadata"
	"net/http"
)

func MetadataMiddleware(ctx context.Context, r *http.Request) metadata.MD {
	// userID 메타데이터로 추가
	md, ok := metadata.FromOutgoingContext(r.Context())
	if ok {
		var kv []string
		if len(md.Get("userID")) > 0 {
			kv = append(kv, "userID", md.Get("userID")[0])
		}

		if len(md.Get("ip")) > 0 {
			kv = append(kv, "ip", md.Get("ip")[0])
		}

		if len(md.Get("accessToken")) > 0 {
			kv = append(kv, "accessToken", md.Get("accessToken")[0])
		}

		if len(md.Get("userAgent")) > 0 {
			kv = append(kv, "userAgent", md.Get("userAgent")[0])
		}

		if len(md.Get("sessionID")) > 0 {
			kv = append(kv, "sessionID", md.Get("sessionID")[0])
		}

		return metadata.Pairs(kv...)
	}

	return nil
}

func InterceptMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newCtx := metadata.AppendToOutgoingContext(r.Context(), "ip", getClientIP(r), "userAgent", r.Header.Get("User-Agent"))
		r = r.WithContext(newCtx)

		next.ServeHTTP(w, r)
	})
}
