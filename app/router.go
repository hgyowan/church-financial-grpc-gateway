package app

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	userV1 "github.com/hgyowan/church-financial-account-grpc/gen/user/v1"
	"github.com/hgyowan/church-financial-grpc-gateway/internal"
	"github.com/hgyowan/church-financial-grpc-gateway/middleware"
	"github.com/hgyowan/go-pkg-library/envs"
	pkgError "github.com/hgyowan/go-pkg-library/error"
	pkgGrpc "github.com/hgyowan/go-pkg-library/grpc-library/grpc"
	pkgLogger "github.com/hgyowan/go-pkg-library/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strings"
)

type router struct {
	mux *runtime.ServeMux
}

func MustNewRouter(ctx context.Context, mux *runtime.ServeMux) *router {
	r := &router{mux: mux}
	if err := r.addHandlerEndpoints(ctx); err != nil {
		pkgLogger.ZapLogger.Logger.Panic(err.Error())
	}
	return r
}

func (r *router) addHandlerEndpoints(ctx context.Context) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := userV1.RegisterUserServiceHandlerFromEndpoint(ctx, r.mux, envs.CFMAccountGRPC, opts); err != nil {
		return pkgError.Wrap(err)
	}

	return nil
}

func (r *router) RegisterHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if path == "/swagger" {
			http.ServeFile(res, req, "./swagger/index.html")
			return
		}

		if path == "/healthz" {
			res.WriteHeader(http.StatusOK)
			return
		}

		if internal.IsSwaggerFile(path) {
			http.ServeFile(res, req, fmt.Sprintf("./swagger%s", path))
			return
		}

		if strings.HasPrefix(path, "/v1/public") {
			base := pkgGrpc.Chain(
				r.mux,
				middleware.InterceptMetadataMiddleware,
			)
			base.ServeHTTP(res, req)
			return
		}

		base := pkgGrpc.Chain(
			r.mux,
			middleware.ValidTokenMiddleware,
		)
		// 미들웨어 통과 시 다음 핸들러 호출
		base.ServeHTTP(res, req)
	})
}
