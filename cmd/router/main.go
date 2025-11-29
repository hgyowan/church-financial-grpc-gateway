package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gatewayfile "github.com/black-06/grpc-gateway-file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hgyowan/church-financial-grpc-gateway/app"
	"github.com/hgyowan/church-financial-grpc-gateway/middleware"
	"github.com/hgyowan/go-pkg-library/envs"
	pkgLogger "github.com/hgyowan/go-pkg-library/logger"
	pkgTrace "github.com/hgyowan/go-pkg-library/trace"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
)

func main() {
	// OpenTelemetry Tracer 초기화
	pkgLogger.MustInitZapLogger()

	bCtx, cancelFunc := context.WithCancel(context.Background())
	group, gCtx := errgroup.WithContext(bCtx)
	doneChan := make(chan struct{}, 1)

	if envs.ServiceType == envs.PrdType {
		shutdown := pkgTrace.InitTracer(gCtx, &pkgTrace.OpenTelemetryConfig{
			ServiceName: envs.ServerName,
			Endpoint:    envs.OpenTelemetryEndpoint,
		})
		defer shutdown()
	}

	gMux := runtime.NewServeMux(
		runtime.WithMetadata(middleware.MetadataMiddleware),
		runtime.WithForwardResponseRewriter(middleware.ResponseEnvelope),
		runtime.WithErrorHandler(middleware.CustomHttpError),
		gatewayfile.WithFileIncomingHeaderMatcher(),
		gatewayfile.WithFileForwardResponseOption(),
		gatewayfile.WithHTTPBodyMarshaler(),
	)

	r := app.MustNewRouter(gCtx, gMux)
	c := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "UPDATE"},
		AllowedHeaders:   []string{"X-Forwarded-For", "X-Request-Id", "X-Forwarded-Proto", "X-Forwarded-Host", "Origin", "Content-Length", "Access-Control-Allow-Origin", "Content-Type", "Accept-Encoding", "X-Requested-With", "X-CSRF-Token", "Cache-Control", "x-user-token", "Baggage"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Cache-Control", "Content-Language", "Content-Type"},
		MaxAge:           int((12 * time.Hour).Seconds()),
		AllowedOrigins:   []string{"http://localhost:3000", envs.CFMHost},
	}).Handler(r.RegisterHandler(gCtx))

	group.Go(func() error {
		s := &http.Server{
			Addr:    fmt.Sprintf(":%s", envs.ServerPort),
			Handler: c,
		}
		err := s.ListenAndServe()
		pkgLogger.ZapLogger.Logger.Info("GRPC Gateway End")
		doneChan <- struct{}{}
		return err
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer close(interrupt)

	select {
	case <-doneChan:
		cancelFunc()
	case <-interrupt:
		cancelFunc()
	}

	if err := group.Wait(); err != nil {
		pkgLogger.ZapLogger.Logger.Fatal(err.Error())
	}

	pkgLogger.ZapLogger.Logger.Info("GRPC Gateway End")
}
