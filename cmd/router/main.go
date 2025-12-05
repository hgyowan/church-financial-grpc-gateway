package main

import (
	"context"
	"errors"
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
	pkgLogger.MustInitZapLogger()

	bCtx, cancelFunc := context.WithCancel(context.Background())
	group, gCtx := errgroup.WithContext(bCtx)

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

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", envs.ServerPort),
		Handler: c,
	}

	group.Go(func() error {
		err := s.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			pkgLogger.ZapLogger.Logger.Info("server closed gracefully")
			return nil
		}
		return err
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer close(interrupt)

	select {
	case <-interrupt:
		pkgLogger.ZapLogger.Logger.Info("received shutdown signal")

		cancelFunc()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			pkgLogger.ZapLogger.Logger.Error("server shutdown failed: " + err.Error())
		} else {
			pkgLogger.ZapLogger.Logger.Info("server gracefully stopped")
		}
	}

	if err := group.Wait(); err != nil {
		pkgLogger.ZapLogger.Logger.Fatal(err.Error())
	}

	pkgLogger.ZapLogger.Logger.Info("GRPC Gateway End")
}
