package middleware

import (
	"context"
	"encoding/json"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pkgError "github.com/hgyowan/go-pkg-library/error"
	pkgLogger "github.com/hgyowan/go-pkg-library/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type Resp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func CustomHttpError(_ context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	pkgLogger.ZapLogger.Logger.Error(err.Error())

	castedErr, ok := pkgError.CastBusinessError(err)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(Resp{
			Code: int(pkgError.None),
			Data: nil,
		})
		if _, err := io.WriteString(w, string(b)); err != nil {
			pkgLogger.ZapLogger.Logger.Sugar().Errorf("failed to wirte response: %v", err)
		}
		return
	}

	buf, err := marshaler.Marshal(Resp{
		Code: castedErr.Status.Code,
		Data: castedErr.Status.Data,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(Resp{
			Code: int(pkgError.None),
			Data: nil,
		})
		if _, err := io.WriteString(w, string(b)); err != nil {
			pkgLogger.ZapLogger.Logger.Sugar().Errorf("failed to wirte response: %v", err)
		}
		return
	}

	defaultErrorStatusCode := http.StatusInternalServerError
	if castedErr.Status.HttpStatusCode != 0 {
		defaultErrorStatusCode = castedErr.Status.HttpStatusCode
	}
	w.WriteHeader(defaultErrorStatusCode)

	startTime := time.Now().UTC()
	entries := make([]zap.Field, 0)
	entries = append(entries, zap.String("method", r.Method))
	entries = append(entries, zap.String("path", r.RequestURI))
	entries = append(entries, zap.Int64("latency_ms", time.Since(startTime).Milliseconds()))
	entries = append(entries, zap.Int("status_code", defaultErrorStatusCode))
	entries = append(entries, zap.Int("custom_code", castedErr.Status.Code))
	entries = append(entries, zap.Any("detail", castedErr.Status.Detail))

	msg := castedErr.Status.Message
	if defaultErrorStatusCode >= 400 && castedErr.Status.Code != 0 {
		pkgLogger.ZapLogger.Logger.Error(msg, entries...)
	} else {
		pkgLogger.ZapLogger.Logger.Info(msg, entries...)
	}

	if _, err := w.Write(buf); err != nil {
		pkgLogger.ZapLogger.Logger.Sugar().Errorf("failed to write response buf: %v", err)
	}
}
