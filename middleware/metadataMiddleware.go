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
		if len(md.Get("user_id")) > 0 {
			kv = append(kv, "user_id", md.Get("user_id")[0])
		}

		if len(md.Get("ip")) > 0 {
			kv = append(kv, "ip", md.Get("ip")[0])
		}

		if len(md.Get("access_token")) > 0 {
			kv = append(kv, "access_token", md.Get("access_token")[0])
		}

		if len(md.Get("user_agent")) > 0 {
			kv = append(kv, "user_agent", md.Get("user_agent")[0])
		}

		return metadata.Pairs(kv...)
	}

	return nil
}

func InterceptMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newCtx := metadata.AppendToOutgoingContext(r.Context(), "ip", getClientIP(r), "user_agent", r.Header.Get("User-Agent"))
		r = r.WithContext(newCtx)

		next.ServeHTTP(w, r)
	})
}
